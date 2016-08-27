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
	if err := saveEndpoint(MASTER, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to register master endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to register master endpoint"})
		return http.StatusInternalServerError, result
	}

	result, _ = json.Marshal(map[string]string{"message": "Successed to register master endpoint"})
	return http.StatusOK, result
}

func GetSynMasterHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	masterlist, err := getSynList(MASTER)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get master endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get master endpoint"})
		return http.StatusInternalServerError, result
	} else if masterlist == "" {
		synlog.Error("[REGISTRY API] No endpoint in the master region")

		result, _ = json.Marshal(map[string]string{"message": "No endpoint in the master region"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(masterlist)
}

func DelSynMasterHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if exists, err := delEndpoint(MASTER, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete master endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete master endpoint"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete master: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	info := fmt.Sprintf("Successed to delete master endpoint")
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func GetTagsListHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil {
		synlog.Error("[REGISTRY API] Failed to get repository %v/%v: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get repository info"})
		return http.StatusBadRequest, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Not found repository %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Not found repository"})
		return http.StatusNotFound, result
	}

	tagslist := r.GetTagslist()
	body, err := json.Marshal(tagslist)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get tag list: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get tag list"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, body
}

func PostSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if err := saveEndpoint(DRC, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to register DRC endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to register DRC endpoint"})
		return http.StatusInternalServerError, result
	}

	result, _ = json.Marshal(map[string]string{"message": "Successed to register DRC endpoint"})
	return http.StatusOK, result
}

func GetSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	drclist, err := getSynList(DRC)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get DRC endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get DRC endpoint"})
		return http.StatusInternalServerError, result
	} else if drclist == "" {
		synlog.Error("[REGISTRY API] No endpoint in the DRC endpoint")

		result, _ = json.Marshal(map[string]string{"message": "No endpoint in the DRC endpoint"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(drclist)
}

func DelSynDRCHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	body, _ := ctx.Req.Body().Bytes()
	if exists, err := delEndpoint(DRC, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete DRC endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete DRC endpoint"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete DRC: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	result, _ = json.Marshal(map[string]string{"message": "Successed to delete DRC endpoint"})
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

	//save region endpoint
	body, _ := ctx.Req.Body().Bytes()
	if err := saveRegionEndpoint(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to register region endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to register region endpoint"})
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

	if err := trigRegionEndpoint(region, auth); err != nil {
		synlog.Error("[REGISTRY API] Failed to synchron: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, []byte{}
}

func GetSynRegionHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	eplist, err := getRegionEndpoint(namespace, repository, tag)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get region endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get region endpoint"})
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
	if exists, err := delRegionEndpoint(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to delete region endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to delete region endpoint"})
		return http.StatusInternalServerError, result
	} else if !exists {
		synlog.Error("[REGISTRY API] Failed to delete region: not found endpoint")

		result, _ = json.Marshal(map[string]string{"message": "Not found endpoint"})
		return http.StatusNotFound, result
	}

	info := fmt.Sprintf("Successed to delete %s/%s:%s endpoint", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func PutSynMetaDataHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	body, _ := ctx.Req.Body().Bytes()
	if err := saveSynMetaDate(namespace, repository, tag, body); err != nil {
		synlog.Error("[REGISTRY API] Failed to save synchron content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	info := fmt.Sprintf("Successed to synchronize %s/%s:%s", namespace, repository, tag)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func PutSynLayerHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte
	digest := ctx.Params(":digest")

	//body, _ := ctx.Req.Body().Bytes()
	if err := saveSynLayer(ctx, digest); err != nil {
		synlog.Error("[REGISTRY API] Failed to save synchron image content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to synchron"})
		return http.StatusInternalServerError, result
	}

	info := fmt.Sprintf("Successed to synchronize images %s/%s:%s", digest)
	result, _ = json.Marshal(map[string]string{"message": info})
	return http.StatusOK, result
}

func GetSynContHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	sc := new(Syncont)
	sc.Layers = make(map[string][]byte)
	if err := fillSynContent(namespace, repository, tag, sc); err != nil {
		synlog.Error("[REGISTRY API] Failed to get synchron content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get synchron"})
		return http.StatusInternalServerError, result
	}

	body, err := json.Marshal(sc)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to fill synchron content: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get synchron content"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, body
}

func PostSynRecHandler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	if err := recoveryCont(namespace, repository, tag); err != nil {
		synlog.Error("[REGISTRY API] Failed to recover local data: %s", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to recover local data"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, []byte("")
}

func GetSynStateHandler(ctx *macaron.Context) (int, []byte) {
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	epstalist, err := getSynchState(namespace, repository, tag)
	if err != nil {
		synlog.Error("[REGISTRY API] Failed to get region endpoint: %s", err.Error())

		result, _ = json.Marshal(map[string]string{"message": "Failed to get synchendpoint state"})
		return http.StatusInternalServerError, result
	} else if epstalist == "" {
		synlog.Error("[REGISTRY API] No state in the synchendpoint")

		result, _ = json.Marshal(map[string]string{"message": "No state in the synchendpoint"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, []byte(epstalist)
}
