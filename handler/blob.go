package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/satori/go.uuid"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

func HeadBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	if !strings.Contains(digest, ":") {
		log.Error("[REGISTRY API V2] Invalid digest format %v", digest)

		result, _ := module.ReportError(module.DIGEST_INVALID, digest)
		return http.StatusBadRequest, result
	}

	tarsum := strings.Split(digest, ":")[1]

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		log.Info("[REGISTRY API V2] Failed to get blob %v: %v", tarsum, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, err.Error())
		return http.StatusBadRequest, result
	} else if !exists {
		log.Info("[REGISTRY API V2] Not found blob: %v", tarsum)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, digest)
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(i.Size))

	return http.StatusOK, []byte("")
}

func PostBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)
	from := ctx.Query("from")
	mount := ctx.Query("mount")

	var name string
	if namespace == "" {
		name = repository
	} else {
		name = namespace + "/" + repository
	}

	if name != from && from != "" && mount != "" {
		tarsum := strings.Split(mount, ":")[1]
		i := new(models.Image)
		if exists, err := i.Get(tarsum); err != nil {
			log.Error("[REGISTRY API V2] Failed to get blob " + tarsum + ": " + err.Error())

			result, _ := module.ReportError(module.UNKNOWN, err.Error())
			return http.StatusInternalServerError, result
		} else if exists {
			random := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", u.Scheme, u.Host, name, mount)
			ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
			ctx.Resp.Header().Set("Docker-Content-Digest", mount)
			ctx.Resp.Header().Set("Location", random)

			return http.StatusCreated, []byte("")
		}
	}

	uuid := utils.MD5(uuid.NewV4().String())
	state := utils.MD5(fmt.Sprintf("%s/%s", name, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/%s?_state=%s", u.Scheme, u.Host, name, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", "0-0")

	return http.StatusAccepted, []byte("")
}

func PatchBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	var name string
	if namespace == "" {
		name = repository
	} else {
		name = namespace + "/" + repository
	}

	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	imagePathTmp := module.GetImagePath(uuid, setting.APIVERSION_V2)
	layerPathTmp := module.GetLayerPath(uuid, "layer", setting.APIVERSION_V2)

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	if !utils.IsDirExist(imagePathTmp) {
		os.MkdirAll(imagePathTmp, os.ModePerm)
	}

	if _, err := os.Stat(layerPathTmp); err == nil {
		os.Remove(layerPathTmp)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(layerPathTmp, data, 0777); err != nil {
		log.Error("[REGISTRY API V2] Failed to save layer %v: %v", layerPathTmp, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, err.Error())
		return http.StatusInternalServerError, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%s", name, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/%s?_state=%s", u.Scheme, u.Host, name, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", fmt.Sprintf("0-%v", len(data)-1))

	return http.StatusAccepted, []byte("")
}

func PutBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	var name string
	if namespace == "" {
		name = repository
	} else {
		name = namespace + "/" + repository
	}

	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	digest := ctx.Query("digest")
	if !strings.Contains(digest, ":") {
		log.Error("[REGISTRY API V2] Invalid digest format %v", digest)

		result, _ := module.ReportError(module.DIGEST_INVALID, digest)
		return http.StatusBadRequest, result
	}
	tarsum := strings.Split(digest, ":")[1]

	imagePathTmp := module.GetImagePath(uuid, setting.APIVERSION_V2)
	layerPathTmp := module.GetLayerPath(uuid, "layer", setting.APIVERSION_V2)
	imagePath := module.GetImagePath(tarsum, setting.APIVERSION_V2)
	layerPath := module.GetLayerPath(tarsum, "layer", setting.APIVERSION_V2)

	reqbody, _ := ctx.Req.Body().Bytes()
	layerlen, err := module.SaveLayerLocal(imagePathTmp, layerPathTmp, imagePath, layerPath, reqbody)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to save layer %v: %v", layerPath, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, err.Error())
		return http.StatusInternalServerError, result
	}

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	i := new(models.Image)
	i.Path, i.Size = layerPath, int64(layerlen)
	if err := i.Save(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Failed to save blob description %v: %v", tarsum, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, err.Error())
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/blobs/%s", u.Scheme, u.Host, name, digest)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusCreated, []byte("")
}

func GetBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	if !strings.Contains(digest, ":") {
		log.Error("[REGISTRY API V2] Invalid digest format %v", digest)

		result, _ := module.ReportError(module.DIGEST_INVALID, digest)
		return http.StatusBadRequest, result
	}

	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Failed to get blob %v: %v", tarsum, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, err)
		return http.StatusBadRequest, result
	} else if !exists {
		log.Error("[REGISTRY API V2] Not found blob: %v", tarsum)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, digest)
		return http.StatusNotFound, result
	}

	layer, err := module.DownloadLayer(i.Path)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get layer %v: %v", i.Path, err.Error())

		result, _ := module.ReportError(module.BLOB_UNKNOWN, digest)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(layer)))

	return http.StatusOK, layer
}

func DeleteBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	if !strings.Contains(digest, ":") {
		log.Error("[REGISTRY API V2] Invalid digest format %v", digest)

		result, _ := module.ReportError(module.DIGEST_INVALID, digest)
		return http.StatusBadRequest, result
	}

	tarsum := strings.Split(digest, ":")[1]
	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Failed to get blob %v: %v", tarsum, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		log.Error("[REGISTRY API V2] Not found blob %v", digest)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, digest)
		return http.StatusNotFound, result
	}

	if i.Count == 0 {
		if err := i.Delete(tarsum); err != nil {
			log.Error("[REGISTRY API V2] Failed to delete blob %v: %v", tarsum, err.Error())

			result, _ := module.ReportError(module.UNKNOWN, err.Error())
			return http.StatusInternalServerError, result
		}
		if err := module.DeleteLayer(tarsum, "layer", setting.APIVERSION_V2); err != nil {

			result, _ := module.ReportError(module.UNKNOWN, err.Error())
			return http.StatusNotFound, result
		}
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", "0")

	return http.StatusAccepted, []byte("")
}
