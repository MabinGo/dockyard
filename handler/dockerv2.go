/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/updateservice"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/signature"
	UUID "github.com/containerops/dockyard/utils/uuid"
	"github.com/containerops/dockyard/utils/validate"
)

// @Title Ping dockyard
// @Description Check that the endpoint implements Docker Registry API V2.
// @Accept json
// @Attention
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Router /v2 [get]
// @ResponseHeaders Content-Type: application/json
// @ResponseHeaders Docker-Distribution-Api-Version: registry/2.0
func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// @Title Get repository name list
// @Description Retrieve a sorted, json list of repositories available in the dockyard.
// @Accept json
// @Attention
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {object} models.Repolist "docker repository json information, eg:{"repositories": [<name>,...]}"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/_catalog [get]
func GetCatalogV2Handler(ctx *macaron.Context) (int, []byte) {
	r := new(models.DockerV2)
	results := []models.DockerV2{}

	if err := r.List(&results); err != nil {
		message := fmt.Sprintf("Failed to list repositories")
		log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	rl := new(models.Repolist)
	for _, v := range results {
		var name string
		if v.Namespace == "" {
			name = v.Repository
		} else {
			name = v.Namespace + "/" + v.Repository
		}
		rl.Repositories = append(rl.Repositories, name)
	}

	result, err := json.Marshal(rl)
	if err != nil {
		message := fmt.Sprintf("Failed to marshal appname")
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, result
}

// @Title Head blob store
// @Description The existence of a layer can be checked via the HEAD request to the blob store API.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param digest path string true "hash of image's layer, standard sha256 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha256:XXX"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 404 {string} string "not found blob, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/blobs/{digest} [head]
// @ResponseHeaders Content-Type: application/json
// @ResponseHeaders Content-Type: application/octet-stream
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Content-Length: <length>
func HeadBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	i := new(models.DockerImageV2)
	i.BlobSum = tarsum
	if exists, err := i.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get blob %s", tarsum)
		log.Warnf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.DIGEST_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found blob: %s", tarsum)
		log.Infof("[REGISTRY API V2] %s", message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(i.Size))

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// @Title Start an upload
// @Description All layer uploads use two steps to manage the upload process, this is first step.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 201 {string} string ""
// @Success 202 {string} string ""
// @Router /v2/{namespace}/{repository}/blobs/uploads [post]
// @ResponseHeaders Content-Type: text/plain; charset=utf-8
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Docker-Upload-Uuid: <uuid>
// @ResponseHeaders Location: /v2/<name>/blobs/uploads/<uuid>
// @ResponseHeaders Range: 0-0
func PostBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	from := ctx.Query("from")
	mount := ctx.Query("mount")
	u := module.NewURLFromRequest(ctx.Req.Request)

	if mount != "" && !validate.IsDigestValid(mount) {
		detail := fmt.Sprintf("%s", mount)
		result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid digest format", detail)
		return http.StatusBadRequest, result
	}

	name := namespace + "/" + repository
	uuid, _ := UUID.NewUUID()
	uuid = utils.MD5(uuid)
	state := utils.MD5(fmt.Sprintf("%s/%d", name, time.Now().UnixNano()/int64(time.Millisecond)))

	result, _ := json.Marshal(map[string]string{})

	if name != from && from != "" && mount != "" {
		random := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", u.Scheme, u.Host, from, mount)
		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Resp.Header().Set("Docker-Content-Digest", mount)
		ctx.Resp.Header().Set("Location", random)

		return http.StatusCreated, result
	}
	random := fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/%s?_state=%s", u.Scheme, u.Host, name, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", "0-0")

	return http.StatusAccepted, result
}

// @Title Upload content of image layer
// @Description Upload a chunk of data for the specified upload.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param uuid path string true "a uuid identifying the upload, eg: XXX?_state=YYY"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param requestbody body string true "blob binary data"
// @Success 202 {string} string ""
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/blobs/uploads/{uuid} [patch]
// @ResponseHeaders Content-Type: text/plain; charset=utf-8
// @ResponseHeaders Docker-Upload-Uuid: <uuid>
// @ResponseHeaders Location: /v2/<name>/blobs/uploads/<uuid>
// @ResponseHeaders Range: 0-0
func PatchBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	name := namespace + "/" + repository
	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	imagePathTmp := module.GetImagePath(uuid, setting.DOCKERAPIV2)
	layerPathTmp := module.GetLayerPath(uuid, "layer", setting.DOCKERAPIV2)

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	if !utils.IsDirExist(imagePathTmp) {
		os.MkdirAll(imagePathTmp, os.ModePerm)
	}

	file, err := os.OpenFile(layerPathTmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		message := fmt.Sprintf("Create tmp File Error.")
		log.Error("[REGISTRY API V2] Create tmp file error: %s %s", layerPathTmp, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}
	defer file.Close()
	size, err := io.Copy(file, ctx.Req.Request.Body)
	if err != nil {
		message := fmt.Sprintf("Failed to save blob %s", layerPathTmp)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%d", name, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/%s?_state=%s", u.Scheme, u.Host, name, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", fmt.Sprintf("0-%v", size-1))

	result, _ := json.Marshal(map[string]string{})
	return http.StatusAccepted, result
}

// @Title Complete the push
// @Description For an upload to be considered complete, the client must submit a PUT request on the upload endpoint with a digest parameter.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param uuid path string true "a uuid identifying the upload, eg: XXX?_status=YYY"
// @Param digest query string true "hash of image's layer, standard sha256 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha256:XXX"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 201 {string} string ""
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/blobs/uploads/{uuid} [put]
// @ResponseHeaders Content-Type: text/plain; charset=utf-8
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Location: /v2/<name>/blobs/uploads/<digest>
func PutBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	name := namespace + "/" + repository
	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	digest := ctx.Query("digest")
	if !validate.IsDigestValid(digest) {
		detail := fmt.Sprintf("%s", digest)
		result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid digest format", detail)
		return http.StatusBadRequest, result
	}
	tarsum := strings.Split(digest, ":")[1]

	imagePathTmp := module.GetImagePath(uuid, setting.DOCKERAPIV2)
	layerPathTmp := module.GetLayerPath(uuid, "layer", setting.DOCKERAPIV2)
	imagePath := module.GetImagePath(tarsum, setting.DOCKERAPIV2)
	layerPath := module.GetLayerPath(tarsum, "layer", setting.DOCKERAPIV2)

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	//lock image table in order to wait for writing
	i := new(models.DockerImageV2)
	i.BlobSum = tarsum
	if err := i.Save(); err != nil {
		if strings.Contains(err.Error(), "source is busy") {
			message := fmt.Sprintf("Failed to save blob description %s", tarsum)
			log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

			result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
			return http.StatusConflict, result
		}
		message := fmt.Sprintf("Failed to save blob description %s", tarsum)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}
	defer func() {
		if err := i.FreeLock(); err != nil {
			panic(err)
		}
	}()

	layerlen, err := module.SaveLayerLocal(imagePathTmp, layerPathTmp, imagePath, layerPath, ctx.Req.Request.Body)
	if err != nil {
		message := fmt.Sprintf("Failed to save layer %s", layerPath)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	i = new(models.DockerImageV2)
	i.BlobSum = tarsum
	if _, err := i.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to save blob description %s", tarsum)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}
	i.Path, i.Size = layerPath, layerlen
	if err := i.Update(); err != nil {
		message := fmt.Sprintf("Failed to save blob description %s", tarsum)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", u.Scheme, u.Host, name, digest)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	result, _ := json.Marshal(map[string]string{})
	return http.StatusCreated, result
}

// @Title Pull the layer
// @Description Retrieve the blob from the registry identified by digest.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param digest path string true "hash of image's layer, standard sha256 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha256:XXX"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 404 {string} string "not found blob, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /v2/{namespace}/{repository}/blobs/{digest} [get]
// @ResponseHeaders Content-Type: application/octet-stream
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Content-Length: <length>
func GetBlobsV2Handler(ctx *macaron.Context) int {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.DockerImageV2)
	i.BlobSum = tarsum
	if available, err := i.Read(); err != nil {
		if strings.Contains(err.Error(), "source is busy") {
			message := fmt.Sprintf("Failed to get blob %s", tarsum)
			log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

			result, _ := module.ReportError(module.DENIED, message, err.Error())
			ctx.Resp.Write(result)
			return http.StatusConflict
		}
		message := fmt.Sprintf("Failed to get blob %s", tarsum)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		ctx.Resp.Write(result)
		return http.StatusInternalServerError
	} else if !available {
		message := fmt.Sprintf("Not found blob: %s", tarsum)
		log.Errorf("[REGISTRY API V2] %s", message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, digest)
		ctx.Resp.Write(result)
		return http.StatusNotFound
	}
	defer func() {
		if err := i.FreeLock(); err != nil {
			panic(err)
		}
	}()

	fd, err := os.Open(i.Path)
	if err != nil {
		message := fmt.Sprintf("Failed to get layer %s", i.Path)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, digest)
		ctx.Resp.Write(result)
		return http.StatusInternalServerError
	}
	defer fd.Close()
	http.ServeContent(ctx.Resp, ctx.Req.Request, tarsum, time.Now(), fd)

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint("%v", i.Size))

	return http.StatusOK
}

// @Title Upload the image manifests
// @Description Once all of the layers for an image are uploaded, the client can upload the image manifest.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param tag path string true "tag of the target manifest, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param Content-Type header string true "media type of manifest"
// @Param requestbody body string true "manifest binary data"
// @Success 201 {string} string ""
// @Success 202 {string} string ""
// @Failure 400 {string} string "bad request, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/manifests/{tag} [put]
// @ResponseHeaders Content-Type: application/octet-stream
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Location: /v2/<name>/manifests/<digest>
func PutManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	name := namespace + "/" + repository
	agent := ctx.Req.Header.Get("User-Agent")
	tag := ctx.Params(":tag")

	manifest, _ := ctx.Req.Body().Bytes()
	digest, err := signature.DigestManifest(manifest)
	if err != nil {
		message := fmt.Sprintf("Failed to get manifest digest")
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusBadRequest, result
	}

	r := new(models.DockerV2)
	r.Namespace, r.Repository = namespace, repository
	condition := new(models.DockerV2)
	*condition = *r
	r.SchemaVersion, r.Agent = "DOCKERAPIV2", agent
	if err := r.Save(condition); err != nil {
		message := fmt.Sprintf("Failed to save repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	err, schema := module.ParseManifest(r.Id, tag, manifest)
	if err != nil {
		message := fmt.Sprintf("Failed to save manifest")
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.MANIFEST_INVALID, message, err.Error())
		return http.StatusBadRequest, result
	}

	var upService us.UpdateService
	fullname := tag
	sha := digest
	if err := upService.Put("dockerv2", namespace, repository, fullname, []string{sha}); err != nil {
		message := fmt.Sprintf("Failed to create a signature for %s/%s/%s", namespace, repository, fullname)
		log.Errorf("%s", message)

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, nil)
		return http.StatusInternalServerError, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		u.Scheme,
		u.Host,
		name,
		digest)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	var status = []int{http.StatusBadRequest, http.StatusAccepted, http.StatusCreated}

	result, _ := json.Marshal(map[string]string{})
	return status[schema], result
}

// @Title Get the list of repository tags
// @Description Fetch the tags under the repository identified by name.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param Content-Type header string true "application/json"
// @Success 200 {object} models.Taglist "all the tags in namespace/repository"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/tags/list [get]
func GetTagsListV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	r := new(models.DockerV2)
	r.Namespace, r.Repository = namespace, repository
	if exists, err := r.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.TAG_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2] %s", message)

		result, _ := module.ReportError(module.TAG_INVALID, message, nil)
		return http.StatusBadRequest, result
	}

	tl := new(models.Taglist)
	tl.Name = fmt.Sprintf("%s/%s", namespace, repository)

	results := []models.DockerTagV2{}
	t := new(models.DockerTagV2)
	t.DockerV2 = r.Id
	if err := t.List(&results); err != nil {
		message := fmt.Sprintf("Failed to get tagslist")
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusBadRequest, result
	}
	for _, v := range results {
		tl.Tags = append(tl.Tags, v.Tag)
	}
	if len(tl.Tags) <= 0 {
		log.Errorf("[REGISTRY API V2] Repository %v/%v tags not found", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Repository tags not found"})
		return http.StatusNotFound, result
	}

	result, _ := json.Marshal(tl)
	return http.StatusOK, result
}

// @Title Get the manifest of image
// @Description Fetch the tags under the repository identified by name.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param tag path string true "tag of the target manifest, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "manifest binary data"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/manifests/{tag} [get]
// @ResponseHeaders Content-Type: <media type of manifest>
// @ResponseHeaders Docker-Content-Digest: <digest>
// @ResponseHeaders Content-Length: <length>
func GetManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	tag := ctx.Params(":tag")

	r := new(models.DockerV2)
	r.Namespace, r.Repository = namespace, repository
	if exists, err := r.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2]%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2]%s", message)

		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, nil)
		return http.StatusBadRequest, result
	}

	t := new(models.DockerTagV2)
	t.DockerV2, t.Tag = r.Id, tag
	if exists, err := t.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get manifest")
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found manifest %s/%s:%s", namespace, repository, tag)
		log.Errorf("[REGISTRY API V2] %s", message)

		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, nil)
		return http.StatusNotFound, result
	}

	digest, err := signature.DigestManifest([]byte(t.Manifest))
	if err != nil {
		message := fmt.Sprintf("Failed to signature manifest")
		log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	contenttype := []string{"", "application/json; charset=utf-8", "application/vnd.docker.distribution.manifest.v2+json"}
	ctx.Resp.Header().Set("Content-Type", contenttype[t.Schema])
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}

// @Title Delete the layer of image
// @Description Delete the blob identified by name and digest.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param digest path string true "hash of image's layer, standard sha256 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha256:XXX"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 404 {string} string "not found blob, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/blobs/{digest} [delete]
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Docker-Content-Digest: <digest>
func DeleteBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]
	i := new(models.DockerImageV2)
	i.BlobSum = tarsum
	if available, err := i.Write(); err != nil {
		if strings.Contains(err.Error(), "source is busy") {
			message := fmt.Sprintf("Failed to delete blob %s", tarsum)
			log.Errorf("[REGISTRY API V2] %s: %v", message, err.Error())

			result, _ := module.ReportError(module.DENIED, message, err.Error())
			return http.StatusConflict, result
		}
		message := fmt.Sprintf("Failed to delete blob %s", tarsum)
		log.Error("[REGISTRY API V2] %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !available {
		message := fmt.Sprintf("Not found blob %v", digest)
		log.Error("[REGISTRY API V2] %s", message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, digest)
		return http.StatusNotFound, result
	}

	if i.Reference == 0 {
		if err := i.Delete(); err != nil {
			message := fmt.Sprintf("Failed to delete blob %s", tarsum)
			log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

			result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
			return http.StatusInternalServerError, result
		}
		imagePath := module.GetImagePath(tarsum, setting.DOCKERAPIV2)
		if err := os.RemoveAll(imagePath); err != nil {
			message := fmt.Sprintf("Failed to delete blob %s", tarsum)
			log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

			result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
			return http.StatusNotFound, result
		}
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", "0")

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// @Title Delete the manifest of image
// @Description Delete the manifest identified by name and digest.
// @Accept json
// @Attention
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param reference path string true "hash of image's layer, standard sha256 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha256:XXX"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 202 {string} string ""
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/manifests/{reference} [delete]
func DeleteManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	//TODO: to consider parallel situation
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	name := namespace + "/" + repository
	reference := ctx.Params(":reference")

	r := new(models.DockerV2)
	r.Namespace, r.Repository = namespace, repository
	if exists, err := r.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository %v/%v", namespace, repository)
		log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.NAME_INVALID, message, detail)
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository %s/%s", namespace, repository)
		log.Errorf("[REGISTRY API V2]%s", message)

		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, detail)
		return http.StatusNotFound, result
	}

	t := new(models.DockerTagV2)
	t.DockerV2 = r.Id
	results := []models.DockerTagV2{}
	if err := t.List(&results); err != nil {
		message := fmt.Sprintf("Failed to get tag list %v/%v", namespace, repository)
		log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.NAME_INVALID, message, nil)
		return http.StatusInternalServerError, result
	}
	//if digest of tag accord with the reference, then delete the tag info
	tagslist := []string{}
	for _, v := range results {
		tagslist = append(tagslist, v.Tag)
	}
	if len(tagslist) <= 0 {
		log.Errorf("[REGISTRY API V2] Repository %v/%v tags not found", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Repository tags not found"})
		return http.StatusNotFound, result
	}
	if err := module.DeleteTagByRefer(r.Id, reference, tagslist); err != nil {
		message := fmt.Sprintf("Failed to delete image")
		log.Error("[REGISTRY API V2] %s: %v", message, err.Error())

		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, err.Error())
		return http.StatusNotFound, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusAccepted, result
}

// @Title Download repository metadata
// @Description You can download the metadata from docker v2 repository.
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "metadata binary data"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/meta [get]
// @ResponseHeaders Content-Type: application/json
// @ResponseHeaders Content-Length: <length>
func GetMetaV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	var upService us.UpdateService
	content, err := upService.ReadMeta("dockerv2", namespace, repository)
	if err != nil {
		message := fmt.Sprintf("Failed to read meta data of %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}

// @Title Download repository metadata signature
// @Description You can download the repository metadata signature from software repository.
// @Param namespace path string true "namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "blob binary data"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /v2/{namespace}/{repository}/metasign [get]
// @ResponseHeaders Content-Type: application/octet-stream
// @ResponseHeaders Content-Length: <length>
func GetMetaSignV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	if err := us.KeyManagerEnabled(); err != nil {
		message := "KeyManager is not enabled or does not set proper"
		log.Errorf("%s: %v", message, err)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	var upService us.UpdateService
	content, err := upService.ReadMetaSign("dockerv2", namespace, repository)
	if err != nil {
		message := fmt.Sprintf("Failed to read meta data of %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}
