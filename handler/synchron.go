package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/synchron"
	//"github.com/containerops/dockyard/utils/setting"
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
	body, _ := ctx.Req.Body().Bytes()
	region := new(models.Region)
	if err := json.Unmarshal(body, &region); err != nil {
		return http.StatusBadRequest, []byte("")
	}
	region.Namespace = namespace
	region.Repository = repository
	region.Tag = tag
	region.Active = true
	if err := region.Save(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to save region: %s/%s:%s", namespace, repository, tag)
		return http.StatusBadRequest, []byte("")
	}
	//TODO: mutex
	synchron.RegionTabs = append(synchron.RegionTabs, *region)

	fmt.Println("####### AddSynRegionHandler 2: ", synchron.RegionTabs)

	return http.StatusOK, []byte("")
}

//执行分发
func PutSynRegionHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	fmt.Println("####### PutSynRegionHandler 0: ")

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	s := new(models.Syn)

	r := new(models.Repository)
	if existed, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V2] Failed to get repository metadata: %v", err.Error())
		return http.StatusBadRequest, []byte("")
	} else if !existed {
		log.Error("[REGISTRY API V2] Not found repository metadata: %s/%s", namespace, repository)
		return http.StatusNotFound, []byte("")
	}
	s.Repository = *r

	t := new(models.Tag)
	if existed, err := t.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Failed to get tag metadata: %v", err.Error())
		return http.StatusBadRequest, []byte("")
	} else if !existed {
		log.Error("[REGISTRY API V2] Not found tag metadata: %s/%s:%s", namespace, repository, tag)
		return http.StatusNotFound, []byte("")
	}
	s.Tag = *t

	//analyze manifest and get image metadate
	tarsumlist, err := module.GetTarsumlist([]byte(t.Manifest))
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get tarsum list: %v", err.Error())
		return http.StatusBadRequest, []byte("")
	}
	//get all images metadata
	for _, v := range tarsumlist {
		i := new(models.Image)
		if exists, err := i.Get(v); err != nil {
			log.Error("[REGISTRY API V2] Failed to get blob %v: %v", v, err.Error())
			return http.StatusBadRequest, []byte("")
		} else if !exists {
			log.Error("[REGISTRY API V2] Not found blob: %v", v)
			return http.StatusNotFound, []byte("")
		}
		s.Images = append(s.Images, *i)
	}

	synchron.SynTabs = append(synchron.SynTabs, *s)

	fmt.Println("####### PutSynRegionHandler 1: ", "tarsumlist", tarsumlist)
	fmt.Println("####### PutSynRegionHandler 2: ", "Repository", synchron.SynTabs[0].Repository)
	fmt.Println("####### PutSynRegionHandler 3: ", "Tag", synchron.SynTabs[0].Tag)
	for i := 0; i < len(synchron.SynTabs[0].Images); i++ {
		fmt.Println("####### PutSynRegionHandler 4: ", "Images", "(", i, ")", "::", synchron.SynTabs[0].Images[i])
	}

	//trigger global synchronous distribution immediately
	//...
	//TODO: some section is not syn correctly if not use docker api

	return http.StatusOK, []byte("")
}
