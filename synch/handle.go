package synch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
)

func PostSynMasterHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if err := SaveContent(MASTER, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to save master content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to save master content"})
		return http.StatusInternalServerError, result
	}

	result, _ = json.Marshal(map[string]string{"message": "Successed to register master synchron info"})
	return http.StatusOK, result
}

func GetSynMasterHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	drclist, err := GetSynList(MASTER)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get master list: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get master list"})
		return http.StatusInternalServerError, result
	} else if drclist == "" {
		synlog.Error("[REGISTRY API] Failed to get master list: No endpoint in the master region")

		result, _ = json.Marshal(map[string]string{"message": "No endpoint in the master region"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(drclist)
}

func DelSynMasterHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if exists, err := DelSynEndpoint(MASTER, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete syn master endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete syn master endpoint"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete syn master: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	info := fmt.Sprintf("Successed to delete master endpoint")
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func PostSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if err := SaveContent(DRC, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to save DRC content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to save DRC content"})
		return http.StatusInternalServerError, result
	}

	result, _ = json.Marshal(map[string]string{"message": "Successed to register DRC synchron info"})
	return http.StatusOK, result
}

func PostSynRegionHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	//search repository whether existed
	t := new(models.Tag)
	if existed, err := t.Get(namespace, repository, tag); err != nil {
		synlog.Error("[REGISTRY API V2] Failed to get tag: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get tag"})
		return http.StatusBadRequest, result
	} else if !existed {
		synlog.Error("[REGISTRY API V2] Not found tag: %s/%s:%s", namespace, repository, tag)

		result, _ = json.Marshal(map[string]string{"message": "Not found tag"})
		return http.StatusNotFound, result
	}

	//get region content
	body, _ := ctx.Req.Body().Bytes()
	if err := SaveRegionContent(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to get region content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to save region content"})
		return http.StatusInternalServerError, result
	}

	info := fmt.Sprintf("Successed to register %s/%s:%s synchron info", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func PostSynTrigHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")
	auth := ctx.Req.Header.Get("Authorization")

	region := new(Region)
	if existed, err := region.Get(namespace, repository, tag); err != nil {
		synlog.Error("[REGISTRY API] Failed to get region: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get region info"})
		return http.StatusInternalServerError, result
	} else if !existed {
		synlog.Error("[REGISTRY API] Not found region")

		result, _ = json.Marshal(map[string]string{"message": "Not found region"})
		return http.StatusNotFound, result
	}

	if err := TrigSynEndpoint(region, auth); err != nil {
		synlog.Error("[REGISTRY API] Failed to synchron: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	info := fmt.Sprintf("Successed to trigger %s/%s:%s synchron", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func PutSynContentHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	body, _ := ctx.Req.Body().Bytes()
	if err := SaveSynContent(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to save syn content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	info := fmt.Sprintf("Successed to synchronize %s/%s:%s", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func GetSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	drclist, err := GetSynList(DRC)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get DRC list: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get DRC list"})
		return http.StatusInternalServerError, result
	} else if drclist == "" {
		synlog.Error("[REGISTRY API] Failed to get DRC list: No endpoint in the DRC region")

		result, _ = json.Marshal(map[string]string{"message": "No endpoint in the DRC region"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(drclist)
}

func GetSynRegionHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	eplist, err := GetSynRegionEndpoint(namespace, repository, tag)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get region info: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get region info"})
		return http.StatusInternalServerError, result
	} else if eplist == "" {
		synlog.Error("[REGISTRY API] No endpoint in the region")

		result, _ = json.Marshal(map[string]string{"message": "No endpoint in the region"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(eplist)
}

func DelSynRegionHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	body, _ := ctx.Req.Body().Bytes()
	if exists, err := DelSynRegion(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete syn region: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete syn region"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete syn region: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	info := fmt.Sprintf("Successed to delete %s/%s:%s endpoint", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func DelSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if exists, err := DelSynEndpoint(DRC, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete syn DRC endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete syn DRC endpoint"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete syn DRC: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	info := fmt.Sprintf("Successed to delete DRC endpoint")
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}
