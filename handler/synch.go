package handler

import (
	//"encoding/json"
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
		return http.StatusBadRequest, []byte("")
	} else if !existed {
		log.Error("[REGISTRY API V2] Not found tag: %s/%s:%s", namespace, repository, tag)
		return http.StatusNotFound, []byte("")
	}

	//get region content
	body, _ := ctx.Req.Body().Bytes()
	if err := module.SaveRegionContent(namespace, repository, tag, body); err != nil {
		log.Error("[REGISTRY API] Failed to get region content: %s", err.Error())
		return http.StatusInternalServerError, []byte("")
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

	if err := module.TrigSynEndpoint(region); err != nil {
		log.Error("[REGISTRY API] Failed to synchronize endpoint: %s", err.Error())
		return http.StatusInternalServerError, []byte("")
	}

	/*
		epg := new(models.Endpointgrp)
		if err := json.Unmarshal([]byte(region.Endpointlist), epg); err != nil {
			log.Error("[REGISTRY API] Failed to unmarshal: %s", err.Error())
			return http.StatusInternalServerError, []byte("")
		}

		success := true
		for k, _ := range epg.Endpoints {
			if epg.Endpoints[k].Active == false {
				continue
			}

			if err := module.TrigSynch(namespace, repository, tag, epg.Endpoints[k].URL); err != nil {
				success = false
				log.Error("[REGISTRY API] Failed to synchronize %s: %s", epg.Endpoints[k].URL, err.Error())
				continue
			} else {
				epg.Endpoints[k].Active = false
			}
		}

		result, _ := json.Marshal(epg)
		region.Endpointlist = string(result)
		if err := region.Save(namespace, repository, tag); err != nil {
			log.Error("[REGISTRY API] Failed to save syn content: %s", err.Error())
			return http.StatusInternalServerError, []byte("")
		}

		if !success {
			return http.StatusInternalServerError, []byte("")
		}
	*/
	return http.StatusOK, []byte("")
}
