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

package module

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils/uuid"
)

var (
	DIGEST_INVALID        = "DIGEST_INVALID"
	SIZE_INVALID          = "SIZE_INVALID"
	NAME_INVALID          = "NAME_INVALID"
	TAG_INVALID           = "TAG_INVALID"
	NAME_UNKNOWN          = "NAME_UNKNOWN"
	MANIFEST_UNKNOWN      = "MANIFEST_UNKNOWN"
	MANIFEST_INVALID      = "MANIFEST_INVALID"
	MANIFEST_UNVERIFIED   = "MANIFEST_UNVERIFIED"
	MANIFEST_BLOB_UNKNOWN = "MANIFEST_BLOB_UNKNOWN"
	BLOB_UNKNOWN          = "BLOB_UNKNOWN"
	BLOB_UPLOAD_UNKNOWN   = "BLOB_UPLOAD_UNKNOWN"
	BLOB_UPLOAD_INVALID   = "BLOB_UPLOAD_INVALID"
	UNKNOWN               = "UNKNOWN"
	UNSUPPORTED           = "UNSUPPORTED"
	UNAUTHORIZED          = "UNAUTHORIZED"
	DENIED                = "DENIED"
	UNAVAILABLE           = "UNAVAILABLE"
	TOOMANYREQUESTS       = "TOOMANYREQUESTS"
	APINOTCOMPATIBLE      = "APINOTCOMPATIBLE"
)

type Errors struct {
	Errors []Errunit `json:"errors"`
}

type Errunit struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

func ReportError(code string, message string, detail interface{}) ([]byte, error) {
	var errs = Errors{}

	item := Errunit{
		Code:    code,
		Message: message,
		Detail:  detail,
	}

	errs.Errors = append(errs.Errors, item)

	return json.Marshal(errs)
}

var Apis = []string{"images", "tarsum", "acis"}

func GetImagePath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v", setting.DockyardPath, Apis[apiversion], imageId)
}

func GetManifestPath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/manifest", setting.DockyardPath, Apis[apiversion], imageId)
}

func GetSignaturePath(imageId, signfile string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/%v", setting.DockyardPath, Apis[apiversion], imageId, signfile)
}

func GetLayerPath(imageId, layerfile string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/%v", setting.DockyardPath, Apis[apiversion], imageId, layerfile)
}

type hmacKey string

var DYHMAC = "dockyard-state"

func GenerateToken(namespace, repository string) (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	s := new(models.State)
	s.Namespace = namespace
	s.Repository = repository
	s.UUID = uuid
	//s.Offset =
	s.Locked = 1
	//s.CreatedAt = 1

	hk := hmacKey(DYHMAC)
	token, err := hk.packUploadState(*s)
	if err != nil {
		return "", err
	}
	fmt.Printf("\n #### mabin 000: token %v \n", err)

	ts := new(models.State)
	ts.Namespace, ts.Repository = namespace, repository
	*ts = *s
	if err := s.Save(ts); err != nil {
		return "", err
	}

	return token, nil
}

func VerifyToken(namespace, repository, token string) error {
	hk := hmacKey(DYHMAC)
	state, err := hk.unpackUploadState(token)
	if err != nil {
		return err
	}

	if (state.Namespace != namespace) || (state.Repository != repository) {
		return fmt.Errorf("invalid repository %v/%v", namespace, repository)
	}

	// TODO:

	return nil
}

func (hk hmacKey) unpackUploadState(token string) (models.State, error) {
	var state models.State

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}
	mac := hmac.New(sha256.New, []byte(hk))

	if len(tokenBytes) < mac.Size() {
		return state, fmt.Errorf("invalid token")
	}

	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, fmt.Errorf("invalid token")
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}

func (hk hmacKey) packUploadState(lus models.State) (string, error) {
	mac := hmac.New(sha256.New, []byte(hk))
	p, err := json.Marshal(lus)
	if err != nil {
		return "", err
	}

	mac.Write(p)
	return base64.URLEncoding.EncodeToString(append(mac.Sum(nil), p...)), nil
}
