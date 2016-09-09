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
	//"crypto/sha512"
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
	"github.com/containerops/dockyard/updateservice"
	"github.com/containerops/dockyard/utils"
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
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/search [get]
func AppGlobalSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	u := ctx.Req.Request.URL

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

	results := []models.ArtifactV1{}
	f := new(models.ArtifactV1)
	if err := f.QueryGlobal(&results, querys...); err != nil {
		message := fmt.Sprintf("Failed to get app")
		log.Errorf("%s: %v", message, err.Error())
		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	httpbodys := []models.SearchOutput{}
	for _, v := range results {
		a := new(models.AppV1)
		a.Id = v.AppV1
		if _, err := a.IsExist(); err != nil {
			message := fmt.Sprintf("Failed to get repository")
			log.Errorf("%s: %v", message, err.Error())

			result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
			return http.StatusInternalServerError, result
		}
		httpbody := models.SearchOutput{
			Namespace:   a.Namespace,
			Repository:  a.Repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         v.URL,
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
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param key query string true "fuzzy search by key follows the url, identify top 4 parameters and separated by '+', eg: url?key=appname+tag"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 404 {string} string "not found namespace/repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/search [get]
func AppScopedSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	u := ctx.Req.Request.URL

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

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

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
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}

	httpbodys := []models.SearchOutput{}
	for _, v := range results {
		httpbody := models.SearchOutput{
			Namespace:   namespace,
			Repository:  repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         v.URL,
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
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {array} models.SearchOutput "application detail info"
// @Failure 400 {string} string "bad request, response error information"
// @Failure 404 {string} string "not found namespace/repository, response error information"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/list [get]
func AppGetListAppV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

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
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusBadRequest, result
	}

	namelists := []models.SearchOutput{}

	for _, v := range results {
		namelists = append(namelists, models.SearchOutput{
			Namespace:   namespace,
			Repository:  repository,
			OS:          v.OS,
			Arch:        v.Arch,
			Name:        v.App,
			Tag:         v.Tag,
			Description: v.Manifests,
			URL:         v.URL,
			Size:        v.Size,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	result, err := json.Marshal(namelists)
	if err != nil {
		message := fmt.Sprintf("Failed to marshal appname")
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	}
	return http.StatusOK, result
}

// @Title Download applications.
// @Description You can download the application from software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application file binary data"
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [get]
// @Response Content Type
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Content-Range: bytes <start>-<end>/<size>
// @ResponseHeaders Content-Type: application/octet-stream
func AppGetFileV1Handler(ctx *macaron.Context) int {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	if err := module.SessionLock(namespace, repository, "pull", setting.APPAPIV1); err != nil {
		fmt.Printf("\n #### mabin AppGetFileV1Handler 000: %v \n", err)
		message := fmt.Sprintf("Failed to get repository file %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.DENIED, message, err.Error())
		ctx.Resp.Write(result)
		return http.StatusConflict
	}
	defer module.SessionUnlock(namespace, repository, "pull", setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	if tag == "" {
		tag = "latest"
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if available, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		ctx.Resp.Write(result)
		return http.StatusInternalServerError
	} else if !available {
		message := fmt.Sprintf("Not found repository or is busy: %s/%s", namespace, repository)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		ctx.Resp.Write(result)
		return http.StatusNotFound
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	if exists, err := i.Read(); err != nil {
		/*
			if strings.Contains(err.Error(), "source is busy") {
				message := fmt.Sprintf("Failed to get app description %s/%s/%s", os, arch, appname)
				log.Errorf("%s: %v", message, err.Error())

				result, _ := module.ReportError(module.DENIED, message, err.Error())
				ctx.Resp.Write(result)
				return http.StatusConflict
			}
		*/
		message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		ctx.Resp.Write(result)
		return http.StatusInternalServerError
	} else if !exists {
		message := fmt.Sprintf("Not found app: %s/%s/%s", system, arch, appname)
		log.Error(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		ctx.Resp.Write(result)
		return http.StatusNotFound
	}

	fd, err := os.Open(i.Path)
	if err != nil {
		message := fmt.Sprintf("Failed to get APP %s", i.Path)
		log.Errorf(message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, err.Error())
		ctx.Resp.Write(result)
		return http.StatusInternalServerError
	}
	defer fd.Close()

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Range", fmt.Sprintf("0-%v", i.Size-1))
	ctx.Resp.Header().Set("Content-Length", fmt.Sprintf("%v", i.Size))

	http.ServeContent(ctx.Resp, ctx.Req.Request, i.BlobSum, time.Now(), fd)

	return http.StatusOK
}

// @Title Download application manifest
// @Description You can download the application manifest of software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application's manifest binary data"
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/manifests/{tag} [get]
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Content-Range: bytes <start>-<end>/<size>
// @ResponseHeaders Content-Type: application/octet-stream
func AppGetManifestsV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	if err := module.SessionLock(namespace, repository, "pull", setting.APPAPIV1); err != nil {
		fmt.Printf("\n #### mabin AppGetManifestsV1Handler 000: %v \n", err)
		message := fmt.Sprintf("Failed to get repository manifest %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.DENIED, message, err.Error())
		return http.StatusConflict, result
	}
	defer module.SessionUnlock(namespace, repository, "pull", setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	if tag == "" {
		tag = "latest"
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

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
		/*
			if strings.Contains(err.Error(), "source is busy") {
				message := fmt.Sprintf("Failed to get app description %s/%s/%s", os, arch, appname)
				log.Errorf("%s: %v", message, err.Error())

				result, _ := module.ReportError(module.DENIED, message, err.Error())
				return http.StatusConflict, result
			}
		*/
		message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return http.StatusInternalServerError, result
	} else if !exists {
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

// @Title Download repository metadata
// @Description You can download the metadata of software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application's metadata binary data"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/meta [get]
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Content-Type: application/json
func AppGetMetaV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	//TODO:
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")

	var upService us.UpdateService
	content, err := upService.ReadMeta("app", namespace, repository)
	if err != nil {
		message := fmt.Sprintf("Failed to read meta data of %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}

// @Title Download repository metadata signature
// @Description You can download the repository metadata signature of software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "application's metadata signature binary data"
// @Failure 500 {string} string "internal server error, response error information of api server"
// @Router /app/v1/{namespace}/{repository}/metasign [get]
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Content-Type: application/octet-stream
func AppGetMetaSignV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	//TODO: auth and check repo validataion
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")

	if err := us.KeyManagerEnabled(); err != nil {
		message := "KeyManager is not enabled or does not set proper"
		log.Errorf("%s: %v", message, err)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	var upService us.UpdateService
	content, err := upService.ReadMetaSign("app", namespace, repository)
	if err != nil {
		message := fmt.Sprintf("Failed to read meta data of %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}

// @Title Request to upload application
// @Description You should request to upload application from software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 202 {string} string ""
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository} [post]
// @ResponseHeaders App-Upload-UUID: <Random UUID>
// @ResponseHeaders Content-Type: text/plain; charset=utf-8
func AppPostV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int

	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")

	sessionid, err := module.GenerateSessionID(namespace, repository, setting.APPAPIV1)
	if err != nil {
		message := fmt.Sprintf("Failed to get App upload UUID %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}

	defer func() {
		fmt.Printf("\n #### mabin AppPostV1Handler 000: respcode=%v \n", respcode)
		if respcode != http.StatusAccepted {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1)
		}
	}()

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	condition := new(models.AppV1)
	*condition = *a
	if err := a.Save(condition); err != nil {
		message := fmt.Sprintf("Failed to save repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("App-Upload-UUID", sessionid)

	respcode = http.StatusAccepted
	result, _ := json.Marshal(map[string]string{})

	return respcode, result
}

// @Title Upload content of application
// @Description You can upload application to software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Param Digest header string true "application's checksum, standard sha512 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha512:a3ed95caeb02..."
// @Param requestbody body string true "application file binary data"
// @Success 201 {string} string ""
// @Failure 400 {string} string "bad request, response error information"
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 409 {string} string "operation is conflicted, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [put]
// @ResponseHeaders App-Upload-UUID: <Random UUID>
// @ResponseHeaders Content-Type: text/plain; charset=utf-8
func AppPutFileV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	defer func() {
		fmt.Printf("\n #### mabin AppPutFileV1Handler 000: respcode=%v \n", respcode)
		if respcode != http.StatusCreated {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1)
		}
	}()

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("%v", err)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, sessionid)
		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	host := ctx.Req.Request.Host
	//authorization := ctx.Req.Header.Get("Authorization")

	digest := ctx.Req.Header.Get("Digest")
	hashes := strings.Split(digest, ":")

	if tag == "" {
		tag = "latest"
	}
	if len(hashes) != 2 {
		message := fmt.Sprintf("Invalid digest format %v", digest)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(module.DIGEST_INVALID, message, digest)
		return respcode, result
	}
	sha := hashes[1]

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		respcode = http.StatusNotFound
		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, sha)
		return respcode, result
	}

	rawurl := fmt.Sprintf("%s://%s/app/v1/%s/%s/%s/%s/%s/%s", setting.ListenMode,
		host, namespace, repository, system, arch, appname, tag)
	imagePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", sha)
	appPath := fmt.Sprintf("%s/%s", imagePath, "app")

	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	condition := new(models.ArtifactV1)
	*condition = *i
	if err := i.Save(condition); err != nil {
		/*
			if strings.Contains(err.Error(), "source is busy") {
				message := fmt.Sprintf("Failed to get app description %s/%s/%s", system, arch, appname)
				log.Errorf("%s: %v", message, err.Error())

				respcode = http.StatusConflict
				result, _ := module.ReportError(module.DENIED, message, err.Error())
				return respcode, result
			}
		*/
		message := fmt.Sprintf("Failed to save app %s", sha)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}

	file, err := os.OpenFile(appPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		log.Error("Create app file error: %s %s", appPath, err.Error())

		respcode = http.StatusBadRequest
		result, _ := json.Marshal(map[string]string{"message": "Create .aci File Error."})
		return respcode, result
	}
	defer file.Close()
	size, err := io.Copy(file, ctx.Req.Request.Body)
	if err != nil {
		message := fmt.Sprintf("Failed to save app %s", appPath)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}

	/*
		sha512h := sha512.New()
		if _, err := file.Seek(0, 0); err != nil {
			message := fmt.Sprintf("Failed to save app %s", appPath)
			log.Errorf("%s: %v", message, err.Error())

			result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
			return http.StatusInternalServerError, result
		}

		if _, err := io.Copy(sha512h, file); err != nil {
			message := fmt.Sprintf("Generate data hash code error %v", err.Error())
			log.Error(message)

			result, _ := module.ReportError(module.DIGEST_INVALID, message, err.Error())
			return http.StatusInternalServerError, result
		}
		hash512 := fmt.Sprintf("%x", sha512h.Sum(nil))

			if isEqual := strings.Compare(sha, hash512); isEqual != 0 {
				message := fmt.Sprintf("App hash is not equel digest %v:%v", hash512, digest)
				log.Error(message)

				result, _ := module.ReportError(module.DIGEST_INVALID, message, digest)
				return http.StatusBadRequest, result
			}
	*/
	var upService us.UpdateService
	fullname := fmt.Sprintf("%s/%s/%s/%s", system, arch, appname, tag)
	if err := upService.Put("app", namespace, repository, fullname, []string{sha}); err != nil {
		message := fmt.Sprintf("Failed to create a signature for %s/%s/%s", namespace, repository, fullname)
		log.Errorf("%s", message)

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, nil)
		return respcode, result
	}

	i = new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	if _, err := i.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get app description %s", sha)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}
	blobSum := i.BlobSum
	i.BlobSum, i.Path, i.Size, i.URL = sha, appPath, size, rawurl
	if deleteBlob, err := i.UpdateBlob(blobSum); err != nil {
		message := fmt.Sprintf("Failed to save app description %s", sha)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	} else if deleteBlob != "" {
		deletePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", deleteBlob)
		os.RemoveAll(deletePath)
	}

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("App-Upload-UUID", sessionid)

	respcode = http.StatusCreated
	result, _ := json.Marshal(map[string]string{})

	return respcode, result
}

// @Title Upload manifest of application
// @Description You can upload manifest of application to software repository.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Param Digest header string true "application's checksum, standard sha512 hash value, contain numbers,letters,colon and xdigit, length is not less than 32. eg: sha512:a3ed95caeb02..."
// @Param requestbody body string true "application's manifest binary data"
// @Success 201 {string} string ""
// @Failure 404 {string} string "not found repository, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/manifests/{tag} [put]
func AppPutManifestV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	defer func() {
		fmt.Printf("\n #### mabin AppPutManifestV1Handler 000: respcode=%v \n", respcode)
		if respcode != http.StatusCreated {
			module.ReleaseSessionID(namespace, repository, setting.APPAPIV1)
		}
	}()

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("%v", err)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, sessionid)
		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")

	if tag == "" {
		tag = "latest"
	}

	digest := ctx.Req.Header.Get("Digest")

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.MANIFEST_INVALID, message, err.Error())
		return respcode, result
	} else if !exists {
		message := fmt.Sprintf("Not found repository: %s/%s", namespace, repository)
		log.Error(message)

		respcode = http.StatusNotFound
		result, _ := module.ReportError(module.MANIFEST_UNKNOWN, message, digest)
		return respcode, result
	}

	manifest, _ := ctx.Req.Body().Bytes()

	i := new(models.ArtifactV1)
	i.AppV1, i.OS, i.Arch, i.App, i.Tag = a.Id, system, arch, appname, tag
	condition := new(models.ArtifactV1)
	*condition = *i
	i.Manifests = string(manifest)
	if err := i.SaveAtom(condition); err != nil {
		message := fmt.Sprintf("Failed to save repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		respcode = http.StatusInternalServerError
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, err.Error())
		return respcode, result
	}

	respcode = http.StatusCreated
	result, _ := json.Marshal(map[string]string{})
	return respcode, result
}

// @Title Update the status of uploading application.
// @Description You can update the uploading status.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param status path string true "uploading status, done or error"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Param App-Upload-UUID header string true "get from software repository and fill in request header"
// @Success 202 {string} string ""
// @Failure 400 {string} string "bad request, parameters or url is error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{status}/{tag} [patch]
func AppPatchFileV1Handler(ctx *macaron.Context) (int, []byte) {
	var respcode int

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	defer module.ReleaseSessionID(namespace, repository, setting.APPAPIV1)

	sessionid := ctx.Req.Header.Get("App-Upload-UUID")
	if err := module.ValidateSessionID(namespace, repository, sessionid, setting.APPAPIV1); err != nil {
		message := fmt.Sprintf("%v", err)
		log.Error(message)

		respcode = http.StatusBadRequest
		result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, sessionid)
		return respcode, result
	}

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	status := ctx.Params(":status")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")

	if strings.EqualFold("done", status) || strings.EqualFold("error", status) {
		respcode = http.StatusAccepted
		result, _ := json.Marshal(map[string]string{})
		return respcode, result
	}

	message := fmt.Sprintf("Failed to upload app %s/%s/%s/%s", system, arch, appname, tag)
	log.Error(message)

	respcode = http.StatusBadRequest
	result, _ := module.ReportError(module.BLOB_UPLOAD_INVALID, message, status)

	return respcode, result
}

// @Title Delete application.
// @Description You can delete application in software package.
// @Accept json
// @Attention
// @Param namespace path string true "application's namespace, only numbers,letters and underscore are allowed, maxlength is 255 byte. eg: Huawei"
// @Param repository path string true "name of application's repository, only numbers,letters,bar and underscore are allowed, maxlength is 255 byte. eg: PaaS"
// @Param os path string true "os type of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: linux"
// @Param arch path string true "architecture of application, non-null, maxlength is 128 byte, input 'undefine' if not sure. eg: amd64"
// @Param app path string true "name of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 255 byte. eg: webapp-v1-linux-amd64.tar.gz"
// @Param tag path string false "tag of application, only numbers,letters,bar,dot and underscore are allowed, maxlength is 128 byte. eg: latest"
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string ""
// @Failure 404 {string} string "not found repository or application, response error information"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /app/v1/{namespace}/{repository}/{os}/{arch}/{app}/{tag} [delete]
func AppDeleteFileV1Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	if err := module.SessionLock(namespace, repository, "delete", setting.APPAPIV1); err != nil {
		fmt.Printf("\n #### mabin AppDeleteFileV1Handler 000: %v \n", err)
		message := fmt.Sprintf("Failed to delete repository file %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.DENIED, message, err.Error())
		return http.StatusConflict, result
	}
	defer module.SessionUnlock(namespace, repository, "delete", setting.APPAPIV1)

	system := ctx.Params(":os")
	arch := ctx.Params(":arch")
	appname := ctx.Params(":app")
	tag := ctx.Params(":tag")
	//host := ctx.Req.Header.Get("Host")
	//authorization := ctx.Req.Header.Get("Authorization")
	if tag == "" {
		tag = "latest"
	}

	a := new(models.AppV1)
	a.Namespace, a.Repository = namespace, repository
	if exists, err := a.IsExist(); err != nil {
		message := fmt.Sprintf("Failed to get repository description %s/%s", namespace, repository)
		log.Errorf("%s: %v", message, err.Error())

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
	if deleteBlob, err := i.Delete(); err != nil {
		if strings.EqualFold(err.Error(), "record not found") {
			message := fmt.Sprintf("Not found app: %s/%s/%s/%s", system, arch, appname, tag)
			log.Error(message)

			result, _ := module.ReportError(module.UNKNOWN, message, nil)
			return http.StatusNotFound, result
		}
		message := fmt.Sprintf("Failed to delete app %s/%s/%s", system, arch, appname)
		log.Errorf("%s: %v", message, err.Error())

		result, _ := module.ReportError(module.UNKNOWN, message, err.Error())
		return http.StatusInternalServerError, result
	} else if deleteBlob != "" {
		deletePath := fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", deleteBlob)
		os.RemoveAll(deletePath)
	}

	var upService us.UpdateService
	fullname := fmt.Sprintf("%s/%s/%s/%s", system, arch, appname, tag)
	if err := upService.Delete("app", namespace, repository, fullname); err != nil {
		message := fmt.Sprintf("Failed to remove signature for %s/%s/%s", namespace, repository, fullname)
		log.Errorf("%s", message)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
