package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/docker/distribution/manifest/schema2"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/synch"
	"github.com/containerops/dockyard/utils/setting"
	"github.com/containerops/dockyard/utils/signature"
)

var ManifestCtx []byte

func PutManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//TODO: to consider parallel situation
	manifest := ManifestCtx
	defer func() {
		ManifestCtx = []byte{}
	}()

	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	u := module.NewURLFromRequest(ctx.Req.Request)

	var name string
	if namespace == "" {
		name = repository
		namespace = "library"
	} else {
		name = namespace + "/" + repository
	}

	agent := ctx.Req.Header.Get("User-Agent")
	tag := ctx.Params(":tag")

	if len(manifest) == 0 {
		manifest, _ = ctx.Req.Body().Bytes()
	}

	t := new(models.Tag)
	//To get the existence of tag; if tag exists, the count of images would not be added
	tagexists, err := t.Get(namespace, repository, tag)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get manifest: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.UNKNOWN, detail)
		return http.StatusBadRequest, result
	}

	digest, err := signature.DigestManifest(manifest)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get manifest digest: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag, "Digest": digest}
		result, _ := module.ReportError(module.DIGEST_INVALID, detail)
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		log.Error("[REGISTRY API V2] Failed to save repository %v/%v: %v", namespace, repository, err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.UNKNOWN, detail)
		return http.StatusInternalServerError, result
	}

	err, schema := module.ParseManifest(manifest, namespace, repository, tag)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to decode manifest: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.MANIFEST_INVALID, detail)
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/manifests/%s",
		u.Scheme,
		u.Host,
		name,
		digest)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	tarsumlist, err := module.GetTarsumlist(manifest)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get tarsum list: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.MANIFEST_INVALID, detail)
		return http.StatusBadRequest, result
	}

	if err := module.UploadLayer(tarsumlist); err != nil {
		log.Error("[REGISTRY API V2] Failed to upload layer: %v", err)

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, detail)
		return http.StatusBadRequest, result
	}

	//to identify whether the same user/repo:tag upload repeatedly
	if tagexists == false {
		if err := module.UpdateImgRefCnt(tarsumlist); err != nil {
			log.Error("[REGISTRY API V2] Failed to update image reference counting: %v", err.Error())

			detail := map[string]string{"Name": name, "Tag": tag}
			result, _ := module.ReportError(module.MANIFEST_BLOB_UNKNOWN, detail)
			return http.StatusBadRequest, result
		}
	}

	if err := module.SaveV2Conversion(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to save v2conversion: " + err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, detail)
		return http.StatusInternalServerError, result
	}

	//TODO:
	//synchronize to DR center
	auth := ctx.Req.Header.Get("Authorization")
	synch.TrigSynDRC(namespace, repository, tag, auth)

	var status = []int{http.StatusBadRequest, http.StatusAccepted, http.StatusCreated}
	return status[schema], []byte("")
}

func GetTagsListV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	var name string
	if namespace == "" {
		name = repository
		namespace = "library"
	} else {
		name = namespace + "/" + repository
	}

	r := new(models.Repository)
	if _, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V2] Failed to get repository %v/%v: %v", namespace, repository, err.Error())

		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.TAG_INVALID, detail)
		return http.StatusBadRequest, result
	}

	data := map[string]interface{}{}

	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	tagslist := r.GetTagslist()
	if len(tagslist) <= 0 {
		log.Error("[REGISTRY API V2] Repository %v/%v tags not found", namespace, repository)

		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.TAG_INVALID, detail)
		return http.StatusNotFound, result
	}
	data["tags"] = tagslist

	result, _ := json.Marshal(data)
	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	acceptHeaders := ctx.Req.Header["Accept"]

	var name string
	if namespace == "" {
		name = repository
		namespace = "library"
	} else {
		name = namespace + "/" + repository
	}

	tag := ctx.Params(":tag")

	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to get manifest: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.UNKNOWN, detail)
		return http.StatusBadRequest, result
	} else if !exists {
		if !synch.IsMasterExisted() {
			log.Error("[REGISTRY API V2] Not found manifest %v/%v:%v", namespace, repository, tag)

			detail := map[string]string{"Name": name, "Tag": tag}
			result, _ := module.ReportError(module.MANIFEST_UNKNOWN, detail)
			return http.StatusNotFound, result
		} else {
			//*******************************************
			auth := ctx.Req.Header.Get("Authorization")
			if err := synch.GetSynFromMaster(namespace, repository, tag, auth); err != nil {
				log.Error("[REGISTRY API V2] Failed to get repository from remote %v/%v:%v", namespace, repository, tag)

				detail := map[string]string{"Name": name, "Tag": tag}
				result, _ := module.ReportError(module.MANIFEST_UNKNOWN, detail)
				return http.StatusNotFound, result
			}
			//*******************************************
		}
	}

	digest, err := signature.DigestManifest([]byte(t.Manifest))
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to signature manifest: %v", err.Error())

		detail := map[string]string{"Name": name, "Tag": tag}
		result, _ := module.ReportError(module.DIGEST_INVALID, detail)
		return http.StatusInternalServerError, result
	}

	manifest := t.Manifest
	supportsSchema2 := false
	for _, mediaType := range acceptHeaders {
		if string(mediaType) == schema2.MediaTypeManifest {
			supportsSchema2 = true
		}
	}
	if !supportsSchema2 && t.Schema == 2 {
		mnf, err := module.ConvertSchema2Manifest(t)
		if err != nil {
			log.Error("[REGISTRY API V2] Failed to conver schema2 manifest: " + err.Error())

			detail := map[string]string{"Name": name, "Tag": tag}
			result, _ := module.ReportError(module.MANIFEST_UNKNOWN, detail)
			return http.StatusInternalServerError, result
		}
		manifest = mnf
	}

	contenttype := []string{"", "application/json; charset=utf-8", "application/vnd.docker.distribution.manifest.v2+json"}
	ctx.Resp.Header().Set("Content-Type", contenttype[t.Schema])
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(manifest)))

	return http.StatusOK, []byte(manifest)
}

func DeleteManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//TODO: to consider parallel situation
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	var name string
	if namespace == "" {
		name = repository
		namespace = "library"
	} else {
		name = namespace + "/" + repository
	}

	reference := ctx.Params(":reference")
	if !strings.Contains(reference, ":") {
		log.Error("[REGISTRY API V2] Invalid reference format %v", reference)

		detail := map[string]string{"Name": name, "Reference": reference}
		result, _ := module.ReportError(module.DIGEST_INVALID, detail)
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V2] Failed to get repository %v/%v: %v", namespace, repository, err.Error())

		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.NAME_INVALID, detail)
		return http.StatusInternalServerError, result
	} else if !exists {
		detail := map[string]string{"Name": name}
		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, detail)
		return http.StatusNotFound, result
	}
	tagslist := r.GetTagslist()
	//if digest of tag accord with the reference, then delete the tag info
	if err := module.DeleteTagByRefer(namespace, repository, reference, tagslist); err != nil {
		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, err.Error())
		return http.StatusNotFound, result
	}

	return http.StatusAccepted, []byte("")
}
