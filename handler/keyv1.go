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
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/updateservice"
)

// @Title Download public key
// @Description You can download the public key from software repository.
// @Accept json
// @Attention
// @Param Host header string false "registry host, eg: Host: containerops.me"
// @Param Authorization header string true "authentication token, fllow the format as <scheme> <token>, eg: Authorization: Bearer token..."
// @Success 200 {string} string "blob binary data"
// @Failure 500 {string} string "internal server error, response error information"
// @Router /key/v1 [get]
// @ResponseHeaders Content-Length: <length>
// @ResponseHeaders Content-Type: application/octet-stream
func KeyGetPublicV1Handler(ctx *macaron.Context) (int, []byte) {
	var upService us.UpdateService
	if err := us.KeyManagerEnabled(); err != nil {
		message := "KeyManager is not enabled or does not set proper"
		log.Errorf("%s: %v", message, err)

		result, _ := module.ReportError(module.UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	mode, err := us.NewKeyManagerMode(setting.KeyManagerMode)
	if err != nil || mode != us.Share {
		message := "Please set KeyManagerMode to 'share' to enable '/key/v1' API"
		log.Errorf("%s", message)

		result, _ := module.ReportError(module.APINOTCOMPATIBLE, message, nil)
		return http.StatusInternalServerError, result
	}

	content, err := upService.ReadPublicKey("", "")
	if err != nil {
		message := "Fail to read public key"
		log.Errorf("%s", message)

		result, _ := module.ReportError(module.BLOB_UNKNOWN, message, nil)
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(content)))

	return http.StatusOK, content
}
