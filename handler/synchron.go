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
		return http.StatusBadRequest, []byte("")
	}

	body, _ := ctx.Req.Body().Bytes()
	if err := json.Unmarshal(body, region); err != nil {
		return http.StatusBadRequest, []byte("")
	}
	fmt.Println("####### PostSynRegionHandler 0: ", *region)
	region.Active = true
	fmt.Println("####### PostSynRegionHandler 1: ", *region)
	if err := region.Save(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to save region: %s/%s:%s", namespace, repository, tag)
		return http.StatusBadRequest, []byte("")
	}
	fmt.Println("####### PostSynRegionHandler 1.5: ", *region)
	//TODO: mutex
	models.Regions = append(models.Regions, *region)

	fmt.Println("####### PostSynRegionHandler 2: ", models.Regions)

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
		return http.StatusBadRequest, []byte("")
	}

	if err := module.SaveSynContent(namespace, repository, tag, sc); err != nil {
		return http.StatusBadRequest, []byte("")
	}

	return http.StatusOK, []byte("")
}

func PostSynTrigHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	fmt.Println("####### PostSynTrigHandler 0: ")

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	region := new(models.Region)
	if existed, err := region.Get(namespace, repository, tag); err != nil {
		return http.StatusBadRequest, []byte("")
	} else if !existed {
		return http.StatusNotFound, []byte("")
	}
	fmt.Println("####### PostSynTrigHandler 1: ", region.Dest)
	if err := module.TrigSyn(namespace, repository, tag, region.Dest); err != nil {
		log.Error("####### PostSynTrigHandler 2: %v", err.Error())
		return http.StatusBadRequest, []byte("")
	}

	return http.StatusOK, []byte("")
}
