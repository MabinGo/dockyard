/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/validate"
)

func AppDiscoveryV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// @Title Search application.
// @Description You can search application by key in whole software repository.
// @Accept json
// @Attention
// @Param key query string true "fuzzy search by key follows the url, identify top 4 parameters and separated by '+', eg: url?key=appname+tag"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 401 {string} string ""
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/search [get]
func AppGlobalSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	u := ctx.Req.Request.URL
	url := module.NewURLFromRequest(ctx.Req.Request)

	values := ctx.Query("key")
	if len(values) == 0 {
		message := fmt.Sprintf("Failed to get query Parameters: %s", u.RawQuery)
		log.Error(message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusBadRequest, result
	}

	//"+" has been changed to " " in url transporting
	keys := strings.Split(values, " ")
	keyslen := len(keys)
	if keyslen == 1 {
		//"+" has been changed to "%2B" in client
		keys = strings.Split(values, "+")
	}
	querys := keys
	if len(keys) > 4 {
		querys = keys[:4]
	}

	for _, v := range querys {
		if !validate.IsCommonValid(v) {
			detail := fmt.Sprintf("%s", v)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid query parameters format", detail)
			return http.StatusBadRequest, result
		}
	}

	results := []models.ArtifactV1{}
	f := new(models.ArtifactV1)
	if err := f.QueryGlobal(&results, querys...); err != nil {
		message := fmt.Sprintf("Failed to get app")
		log.Errorf("%s: %s", message, err.Error())
		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	httpbodys := []models.SearchOutput{}
	for _, v := range results {
		if v.Active != 1 {
			continue
		}

		a := new(models.AppV1)
		a.Id = v.AppV1
		if _, err := a.IsExist(); err != nil {
			message := fmt.Sprintf("Failed to get repository")
			log.Errorf("%s: %s", message, err.Error())

			result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
			return http.StatusInternalServerError, result
		}

		rawurl := fmt.Sprintf("%s://%s/app/v1/%s/%s/%s/%s/%s/%s", url.Scheme,
			url.Host, a.Namespace, a.Repository, v.OS, v.Arch, v.App, v.Tag)

		httpbody := models.SearchOutput{
			Namespace:   a.Namespace,
			Repository:  a.Repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         rawurl,
			Size:        v.Size,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
		httpbodys = append(httpbodys, httpbody)
	}

	result, _ := json.Marshal(&httpbodys)
	return http.StatusOK, result
}

// @Title Search the detail of application.
// @Description You can search detail of applications by key in namespace's repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param key query string true "fuzzy search by key follows the url, identify top 4 parameters and separated by '+', eg: url?key=appname+tag"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found namespace/repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/search [get]
func AppScopedSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	url := module.NewURLFromRequest(ctx.Req.Request)
	req := ctx.Req.Request
	u := req.URL

	values := ctx.Query("key")
	if len(values) == 0 {
		message := fmt.Sprintf("Failed to get query Parameters: %s", u.RawQuery)
		log.Error(message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)

		return http.StatusBadRequest, result
	}

	//"+" has been changed to " " in url transporting
	keys := strings.Split(values, " ")
	keyslen := len(keys)
	if keyslen == 1 {
		//"+" has been changed to "%2B" in client
		keys = strings.Split(values, "+")
	}
	querys := keys
	if len(keys) > 4 {
		querys = keys[:4]
	}
	for _, v := range querys {
		if !validate.IsCommonValid(v) {
			detail := fmt.Sprintf("%s", v)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid query parameters format", detail)

			return http.StatusBadRequest, result
		}
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	results := []models.ArtifactV1{}
	i := &models.ArtifactV1{AppV1: a.Id}
	if err := i.QueryScope(&results, querys...); err != nil {
		message := fmt.Sprintf("Failed to get app")
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	}

	httpbodys := []models.SearchOutput{}
	for _, v := range results {
		if v.Active != 1 {
			continue
		}
		rawurl := fmt.Sprintf("%s://%s/app/v1/%s/%s/%s/%s/%s/%s", url.Scheme,
			url.Host, namespace, repository, v.OS, v.Arch, v.App, v.Tag)

		httpbody := models.SearchOutput{
			Namespace:   namespace,
			Repository:  repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         rawurl,
			Size:        v.Size,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
		httpbodys = append(httpbodys, httpbody)
	}

	result, _ := json.Marshal(&httpbodys)

	return http.StatusOK, result
}

// @Title List all application.
// @Description You can list all the applications in namespace's repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found namespace/repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/list [get]
func AppGetListAppV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	url := module.NewURLFromRequest(ctx.Req.Request)

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	i := new(models.ArtifactV1)
	i.AppV1 = a.Id
	results := []models.ArtifactV1{}
	if err := i.List(&results); err != nil {
		message := fmt.Sprintf("Failed to get app %v", a.Id)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusBadRequest, result
	}

	namelists := []models.SearchOutput{}

	for _, v := range results {
		if v.Active != 1 {
			continue
		}
		rawurl := fmt.Sprintf("%s://%s/app/v1/%s/%s/%s/%s/%s/%s", url.Scheme,
			url.Host, namespace, repository, v.OS, v.Arch, v.App, v.Tag)

		namelists = append(namelists, models.SearchOutput{
			Namespace:   namespace,
			Repository:  repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         rawurl,
			Size:        v.Size,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	result, err := json.Marshal(namelists)
	if err != nil {
		message := fmt.Sprintf("Failed to marshal appname")
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	}

	return http.StatusOK, result
}

// @Title Download applications.
// @Description You can download the application from software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application file binary data"
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [get]
// @Response Content Type
// @ResponseHeaders "Content-Length" "length"
// @ResponseHeaders "Content-Range" "bytes <start>-<end>/<size>"
// @ResponseHeaders "Content-Type" "application/octet-stream"
func AppGetFileV1Handler(ctx *macaron.Context) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		ctx.Resp.WriteHeader(http.StatusBadRequest)
		ctx.Resp.Write(result)
		return
	}

	if err := module.SessionLock(namespace, repository, module.PULL, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to get file")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			result, _ := module.ReportError(module.DENIED, message, err.Error())
			ctx.Resp.WriteHeader(http.StatusConflict)
			ctx.Resp.Write(result)
			return
		}

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		ctx.Resp.Write(result)
		return
	}
	defer module.SessionUnlock(namespace, repository, module.PULL, setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	if tag == "" {
		tag = "latest"
	}
	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		ctx.Resp.WriteHeader(http.StatusBadRequest)
		ctx.Resp.Write(result)
		return
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if available, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		ctx.Resp.Write(result)
		return
	} else if !available {
		message := fmt.Sprintf("Not found repository or is busy: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		ctx.Resp.WriteHeader(http.StatusNotFound)
		ctx.Resp.Write(result)
		return
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	if exists, err := i.Read(); err != nil {
		message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		ctx.Resp.Write(result)
		return
	} else if !exists || i.Active != 1 {
		message := fmt.Sprintf("Not found app: %s/%s/%s", system, arch, appname)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		ctx.Resp.WriteHeader(http.StatusNotFound)
		ctx.Resp.Write(result)
		return
	}

	module.AppFileLock.Lock()
	defer module.AppFileLock.Unlock()
	fd, err := os.Open(i.Path)
	if err != nil {
		message := fmt.Sprintf("Failed to get APP %s", i.Path)
		log.Errorf(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		ctx.Resp.Write(result)
		return
	}
	defer fd.Close()

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Range", fmt.Sprintf("0-%v", i.Size-1))
	ctx.Resp.Header().Set("Content-Length", fmt.Sprintf("%v", i.Size))
	http.ServeContent(ctx.Resp, ctx.Req.Request, i.BlobSum, time.Now(), fd)

	return
}

// @Title Download application manifest
// @Description You can download the application manifest of software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application's manifest binary data"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/manifests/{tag} [get]
// @ResponseHeaders "Content-Length" "length"
// @ResponseHeaders "Content-Range" "bytes <start>-<end>/<size>"
// @ResponseHeaders "Content-Type" "application/octet-stream"
func AppGetManifestsV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	if err := module.SessionLock(namespace, repository, module.PULL, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to get manifest")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			result, _ := module.ReportError(module.DENIED, message, err.Error())
			return http.StatusConflict, result
		}

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}
	defer module.SessionUnlock(namespace, repository, module.PULL, setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	if tag == "" {
		tag = "latest"
	}
	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	if exists, err := i.Read(); err != nil {
		message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists || i.Active != 1 {
		message := fmt.Sprintf("Not found app: %s/%s/%s", system, arch, appname)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	content := []byte(i.Manifests)

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Range", fmt.Sprintf("0-%v", len(content)-1))
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}

// @Title Request to upload application
// @Description You should request to upload application from software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 202 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository} [post]
// @ResponseHeaders "App-Upload-UUID" "Random UUID"
// @ResponseHeaders "Content-Type" "text/plain; charset=utf-8"
func AppPostV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int
	var result []byte

	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)
		return respcode, result
	}

	_, sessionid, err := module.GenerateSessionID(namespace, repository, setting.APPAPIV1)
	if err != nil {
		message := fmt.Sprintf("Failed to get app upload UUID")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			respcode = http.StatusConflict
			result, _ = module.ReportError(module.DENIED, message, err.Error())
		} else {
			respcode = http.StatusInternalServerError
			result, _ = module.ReportError(module.UNKNOWN, message, err.Error())
		}

		return respcode, result
	}

	defer func() {
		if respcode != http.StatusAccepted {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)
		}
	}()

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	condition := new(models.AppV1)
	*condition = *a
	if err := a.Save(condition); err != nil {
		message := fmt.Sprintf("Failed to save repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	}

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("App-Upload-UUID", sessionid)

	respcode = http.StatusAccepted
	result, _ = json.Marshal(map[string]string{})

	return respcode, result
}

// @Title Upload content of application
// @Description You can upload application to software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Param Digest header string true "application's checksum, standard sha512 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha512:a3ed95caeb02..."
// @Param requestbody body string true "application file binary data"
// @Success 201 {string} string ""
// @Failure 400 {string} string "bad request, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 413 {string} string "request entity is too large, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [put]
// @ResponseHeaders "App-Upload-UUID" "Random UUID"
// @ResponseHeaders "Content-Type" "text/plain; charset=utf-8"
func AppPutFileV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)
		return respcode, result
	}

	defer func() {
		if respcode != http.StatusCreated {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)
		}
	}()

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if _, err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to save file")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			respcode = http.StatusConflict
			result, _ = module.ReportError(module.DENIED, message, err.Error())
		} else {
			respcode = http.StatusBadRequest
			result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		}

		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	if tag == "" {
		tag = "latest"
	}
	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)

		return respcode, result
	}

	digest := ctx.Req.Header.Get("Digest")
	if errcode, err := module.ValidateDigest(digest); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)

		return respcode, result
	}
	hashes := strings.Split(digest, ":")
	if len(hashes) != 2 {
		message := fmt.Sprintf("Invalid digest format %s", digest)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ = module.ReportError(module.DIGEST_INVALID, message, digest)

		return respcode, result
	}
	sha := hashes[1]

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		respcode = http.StatusNotFound
		result, _ = module.ReportError(module.BLOB_UNKNOWN, message, sha)

		return respcode, result
	}

	imagePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", sha)
	appPath := fmt.Sprintf("%s/%s", imagePath, "app")

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	condition := new(models.ArtifactV1)
	*condition = *i
	if err := i.Save(condition); err != nil {
		message := fmt.Sprintf("Failed to save app %s", sha)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	}

	module.SaveImageID(namespace, repository, i.Id, setting.APPAPIV1)

	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, 0750)
	}

	module.AppFileLock.Lock()
	defer module.AppFileLock.Unlock()
	file, err := os.OpenFile(appPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		log.Errorf("Create app file error: %s %s", appPath, err.Error())

		respcode = http.StatusBadRequest
		result, _ = json.Marshal(map[string]string{"message": "Create .aci File Error."})

		return respcode, result
	}
	defer file.Close()

	size, err := io.Copy(file, ctx.Req.Request.Body)
	if err != nil {
		message := fmt.Sprintf("Failed to save app %s", appPath)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	}

	if size > setting.MaxUploadFileSize {
		message := fmt.Sprintf("File too large when adding file to app %s", appPath)
		log.Error(message)
		errMsg, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, nil)

		return http.StatusRequestEntityTooLarge, errMsg
	}
	sha512h := sha512.New()
	if _, err := file.Seek(0, 0); err != nil {
		message := fmt.Sprintf("Failed to save app %s", appPath)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	}

	if _, err := io.Copy(sha512h, file); err != nil {
		message := fmt.Sprintf("Generate data hash code error %s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(module.DIGEST_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	}
	hash512 := fmt.Sprintf("%x", sha512h.Sum(nil))

	if isEqual := strings.Compare(sha, hash512); isEqual != 0 {
		message := fmt.Sprintf("App hash is not equel digest %s:%s", hash512, digest)
		log.Error(message)

		result, _ := module.ReportError(module.DIGEST_INVALID, message, digest)

		return http.StatusBadRequest, result
	}

	blobSum := i.BlobSum
	id := i.Id
	i = new(models.ArtifactV1)
	i.Id = id
	if _, err := i.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to save app %s", sha)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	}
	i.BlobSum, i.Path, i.Size = sha, appPath, size
	if deleteBlob, err := i.UpdateBlob(blobSum); err != nil {
		message := fmt.Sprintf("Failed to save app description %s", sha)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	} else if deleteBlob != "" {
		deletePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", deleteBlob)
		os.RemoveAll(deletePath)
	}

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("App-Upload-UUID", sessionid)

	respcode = http.StatusCreated
	result, _ = json.Marshal(map[string]string{})

	return respcode, result
}

// @Title Upload manifest of application
// @Description You can upload manifest of application to software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Param Digest header string true "application's checksum, standard sha512 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha512:a3ed95caeb02..."
// @Param requestbody body string true "application's manifest binary data"
// @Success 201 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 413 {string} string "request entity is too large, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/manifests/{tag} [put]
func AppPutManifestV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)
		return respcode, result
	}

	defer func() {
		if respcode != http.StatusCreated {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)
		}
	}()

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if _, err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to save manifest")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			respcode = http.StatusConflict
			result, _ = module.ReportError(module.DENIED, message, err.Error())
		} else {
			respcode = http.StatusBadRequest
			result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		}

		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	if tag == "" {
		tag = "latest"
	}
	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)
		return respcode, result
	}

	digest := ctx.Req.Header.Get("Digest")
	if errcode, err := module.ValidateDigest(digest); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)

		return respcode, result
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.MANIFEST_INVALID, message, err.Error())

		return respcode, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		respcode = http.StatusNotFound
		result, _ = module.ReportError(module.MANIFEST_UNKNOWN, message, digest)

		return respcode, result
	}

	manifest, err := ctx.Req.Body().Bytes()

	if err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.MANIFEST_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	}

	if len(manifest) > (256 * 1024) { //_256_KIB decleared in handler/webv1.go
		message := fmt.Sprintf("Failed to save the description of %s/%s, manifests too large", namespace, repository)
		log.Error(message)

		errMsg, _ := module.ReportError(module.DENIED, message, nil)

		return http.StatusRequestEntityTooLarge, errMsg
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	condition := new(models.ArtifactV1)
	*condition = *i
	i.Manifests = string(manifest)
	if err := i.SaveAtom(condition); err != nil {
		message := fmt.Sprintf("Failed to save repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return respcode, result
	}

	respcode = http.StatusCreated
	result, _ = json.Marshal(map[string]string{})
	return respcode, result
}

// @Title Update the status of uploading application.
// @Description You can update the uploading status.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param status path string true "uploading status, done or error"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Success 202 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{status}/{tag} [patch]
func AppPatchFileV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int
	var result []byte

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)

		module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)
		return respcode, result
	}

	//defer module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.END)

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if _, err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to patch status")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			respcode = http.StatusConflict
			result, _ = module.ReportError(module.DENIED, message, err.Error())
		} else {
			respcode = http.StatusBadRequest
			result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		}

		module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)

		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	if tag == "" {
		tag = "latest"
	}

	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(errcode, message, nil)

		module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)

		return respcode, result
	}

	status := ctx.Params(":status")
	if (status != "done") && (status != "error") {
		message := fmt.Sprintf("Invalid status %s", status)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(module.UNKNOWN, "Invalid status", nil)

		module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)

		return respcode, result
	}

	if strings.EqualFold("done", status) || strings.EqualFold("error", status) {
		respcode = http.StatusAccepted
		result, _ = json.Marshal(map[string]string{})

		module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.END)

		return respcode, result
	}

	message := fmt.Sprintf("Failed to upload app %s/%s/%s/%s", system, arch, appname, tag)
	log.Error(message)

	respcode = http.StatusBadRequest
	result, _ = module.ReportError(module.BLOB_UPLOAD_INVALID, message, status)

	module.ReleaseSessionID(namespace, repository, setting.APPAPIV1, module.RUNNING)

	return respcode, result
}

// @Title Delete application.
// @Description You can delete application in software package.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefined' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, if empty, it will be set to 'latest', maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [delete]
func AppDeleteFileV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	if err := module.SessionLock(namespace, repository, module.DELETE, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to delete file")
		log.Errorf("%s: %s", message, err.Error())

		if strings.Contains(err.Error(), "source is busy") {
			result, _ := module.ReportError(module.DENIED, message, err.Error())
			return http.StatusConflict, result
		}

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}
	defer module.SessionUnlock(namespace, repository, module.DELETE, setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	if tag == "" {
		tag = "latest"
	}

	if errcode, err := module.ValidateParams(system, arch, appname, tag); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)

		return http.StatusBadRequest, result
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	if exists, err := i.Read(); err != nil {
		message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())

		return http.StatusInternalServerError, result
	} else if !exists || i.Active != 1 {
		message := fmt.Sprintf("Not found app: %s/%s/%s/%s", system, arch, appname, tag)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	if deleteBlob, err := i.Delete(); err != nil {
		message := fmt.Sprintf("Failed to delete app %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %s", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())

		return http.StatusInternalServerError, result
	} else if deleteBlob != "" {
		deletePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", deleteBlob)
		os.RemoveAll(deletePath)
	}

	result, _ := json.Marshal(map[string]string{})

	return http.StatusOK, result
}

// @Title Recycle application resource.
// @Description You can recycle application resource when namespace/repository is locked in exception.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, follow the format as \<scheme\> \<token\>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Failure 401 {string} string ""
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/recycle [delete]
func AppRecycleV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	if errcode, err := module.ValidateName(namespace, repository); err != nil {
		message := fmt.Sprintf("%s", err.Error())
		log.Error(message)

		result, _ := module.ReportError(errcode, message, nil)
		return http.StatusBadRequest, result
	}

	if exists, err := module.RecycleSession(namespace, repository, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("Failed to recycle %s/%s: %s", namespace, repository, err.Error())
		log.Errorf("%s", message)

		result, _ := module.ReportError(module.DENIED, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.NAME_UNKNOWN, message, nil)

		return http.StatusNotFound, result
	}

	result, _ := json.Marshal(map[string]string{})

	return http.StatusOK, result
}
