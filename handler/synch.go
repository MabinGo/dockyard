package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
)

func PostSynRegionHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	//search repository whether existed
	t := new(models.Tag)
	if existed, err := t.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to get tag: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get tag"})
		return http.StatusBadRequest, result
	} else if !existed {
		log.Error("[REGISTRY API V2] Not found tag: %s/%s:%s", namespace, repository, tag)

		result, _ = json.Marshal(map[string]string{"message": "Not found tag"})
		return http.StatusNotFound, result
	}

	//get region content
	body, _ := ctx.Req.Body().Bytes()
	if err := module.SaveRegionContent(namespace, repository, tag, body); err != nil {
		log.Error("[REGISTRY API] Failed to get region content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to save region content"})
		return http.StatusInternalServerError, result
	}

	result = []byte(fmt.Sprintf("[REGISTRY API] Register %s/%s:%s synchron info successfully\n", namespace, repository, tag))
	return http.StatusOK, result
}

func PostSynTrigHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")
	auth := ctx.Req.Header.Get("Authorization")

	region := new(models.Region)
	if existed, err := region.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API] Failed to get region: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get region info"})
		return http.StatusInternalServerError, result
	} else if !existed {
		log.Error("[REGISTRY API] Not found region")

		result, _ = json.Marshal(map[string]string{"message": "Not found region"})
		return http.StatusNotFound, result
	}

	if err := module.TrigSynEndpoint(region, auth); err != nil {
		log.Error("[REGISTRY API] Failed to synchron: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	result = []byte(fmt.Sprintf("[REGISTRY API] Trigger %s/%s:%s synchron successfully\n", namespace, repository, tag))
	return http.StatusOK, result
}

func PutSynContentHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	body, _ := ctx.Req.Body().Bytes()
	if err := module.SaveSynContent(namespace, repository, tag, body); err != nil {
		log.Error("[REGISTRY API] Failed to save syn content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	result = []byte(fmt.Sprintf("[REGISTRY API] Synchronize %s/%s:%s successfully\n", namespace, repository, tag))
	return http.StatusOK, result
}
