package handler

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
)

func PostSynRegionHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	//search repository whether existed
	t := new(models.Tag)
	if existed, err := t.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to get tag: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get tag"})
		return http.StatusBadRequest, result
	} else if !existed {
		log.Error("[REGISTRY API V2] Not found tag: %s/%s:%s", namespace, repository, tag)

		result, _ := json.Marshal(map[string]string{"message": "Not found tag"})
		return http.StatusNotFound, result
	}

	//get region content
	body, _ := ctx.Req.Body().Bytes()
	if err := module.SaveRegionContent(namespace, repository, tag, body); err != nil {
		log.Error("[REGISTRY API] Failed to get region content: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save region content"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, []byte("")
}

func PutSynContentHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	body, _ := ctx.Req.Body().Bytes()
	if err := module.SaveSynContent(namespace, repository, tag, body); err != nil {
		log.Error("[REGISTRY API] Failed to save syn content: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save synchron content"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, []byte("")
}

func PostSynTrigHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	region := new(models.Region)
	if existed, err := region.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API] Failed to get region: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get region info"})
		return http.StatusInternalServerError, result
	} else if !existed {
		log.Error("[REGISTRY API] Not found region")

		result, _ := json.Marshal(map[string]string{"message": "Not found region"})
		return http.StatusNotFound, result
	}

	if err := module.TrigSynEndpoint(region); err != nil {
		log.Error("[REGISTRY API] Failed to synchron: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, []byte("")
}
