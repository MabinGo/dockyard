package handler

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
)

//注册分发区域的目的地
func PostSynRegionHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")
	/*
		//判断要分发的repository是否存在
		t := new(models.Tag)
		if existed, err := t.Get(namespace, repository, tag); err != nil {
			log.Error("[REGISTRY API V2] Failed to get tag: %s", err.Error())
			return http.StatusBadRequest, []byte("")
		} else if !existed {
			log.Error("[REGISTRY API V2] Not found tag: %s/%s:%s", namespace, repository, tag)
			return http.StatusNotFound, []byte("")
		}
	*/

	//get region description
	region := new(models.Region)
	if _, err := region.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API] Failed to get region: %s", err.Error())
		return http.StatusInternalServerError, []byte("")
	}

	body, _ := ctx.Req.Body().Bytes()
	if err := json.Unmarshal(body, region); err != nil {
		log.Error("[REGISTRY API] Failed to get request body: %s", err.Error())
		return http.StatusBadRequest, []byte("")
	}
	region.Active = true

	if err := region.Save(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API] Failed to save region: %s", err.Error())
		return http.StatusInternalServerError, []byte("")
	}
	//TODO: mutex
	models.Regions = append(models.Regions, *region)

	return http.StatusOK, []byte("")
}

func PutSynContentHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	sc := new(models.Syncont)
	sc.Layers = make(map[string][]byte)
	body, _ := ctx.Req.Body().Bytes()
	if err := json.Unmarshal(body, sc); err != nil {
		log.Error("[REGISTRY API] Failed to get request body: %s", err.Error())
		return http.StatusBadRequest, []byte("")
	}

	if err := module.SaveSynContent(namespace, repository, tag, sc); err != nil {
		log.Error("[REGISTRY API] Failed to save syn content: %s", err.Error())
		return http.StatusInternalServerError, []byte("")
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
		return http.StatusInternalServerError, []byte("")
	} else if !existed {
		log.Error("[REGISTRY API] Not found region")
		return http.StatusNotFound, []byte("")
	}

	if err := module.TrigSyn(namespace, repository, tag, region.Dest); err != nil {
		log.Error("[REGISTRY API] Failed to trigger syn: %s", err.Error())
		return http.StatusBadRequest, []byte("")
	}

	return http.StatusOK, []byte("")
}
