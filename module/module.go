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
	"sync"

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

var DYSESSIONID = "dockyardsessionid"

var (
	UNLOCK = 0
	LOCK   = 1
)

var StateLock sync.RWMutex

func SessionLock() {
	StateLock.Lock()
}

func SessionUnlock() {
	StateLock.Unlock()
}

func IsSessionActive(namespace, repository string) (bool, error) {
	SessionLock()
	defer SessionUnlock()
	current := new(models.Session)
	current.Namespace, current.Repository = namespace, repository
	if exists, err := current.Read(); err != nil {
		return false, err
	} else if !exists {
		return false, fmt.Errorf("not found repository %v/%v", namespace, repository)
	} else {

	}

	return true, nil
}

func GetSessionID(namespace, repository string) (string, error) {
	SessionLock()
	defer SessionUnlock()

	hk := hmacKey(DYSESSIONID)
	origin := new(models.Session)
	origin.Namespace, origin.Repository = namespace, repository
	exists, err := origin.Read()
	if err != nil {
		return "", err
	}

	if exists {
		if ori.Locked == LOCK {
			return "", fmt.Errorf("%v/%v is busy", namespace, repository)
		} else {
			origin.Locked = LOCK
			sessionid, err := hk.packUploadState(*origin)
			if err != nil {
				return "", err
			}

			if err := origin.UpdateLockState(LOCK); err != nil {
				return err
			}
			return sessionid, nil
		}
	}

	current := new(models.Session)
	current.Namespace = namespace
	current.Repository = repository
	current.Locked = LOCK

	sessionid, err := hk.packUploadState(*current)
	if err != nil {
		return "", err
	}
	fmt.Printf("\n #### mabin 000: sessionid %v \n", err)

	tmp := new(models.Session)
	tmp.Namespace, tmp.Repository = namespace, repository
	*tmp = *current
	if err := current.Save(tmp); err != nil {
		return "", err
	}

	return sessionid, nil
}

func ValidateSessionID(namespace, repository, sessionid string) error {
	hk := hmacKey(DYSESSIONID)
	session, err := hk.unpackUploadState(sessionid)
	if err != nil {
		return err
	}

	if (session.Namespace != namespace) || (session.Repository != repository) {
		return fmt.Errorf("mismatch repository %v/%v", namespace, repository)
	}

	SessionLock()
	defer SessionUnlock()
	current := new(models.Session)
	current.Namespace, current.Repository = namespace, repository
	if exists, err := current.Read(); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found repository %v/%v", namespace, repository)
	}

	if current.Locked != session.Locked {
		return fmt.Errorf("invalid session lock state %v", session.Locked)
	}

	if current.Locked == LOCK {
		return fmt.Errorf("%v/%v is busy", namespace, repository)
	} else if current.Locked == UNLOCK {
		// TODO:
		if err := current.UpdateLockState(LOCK); err != nil {
			return err
		}
	}

	return nil
}

func (hk hmacKey) unpackUploadState(sessionid string) (models.Session, error) {
	var session models.Session

	content, err := base64.URLEncoding.DecodeString(sessionid)
	if err != nil {
		return models.Session{}, err
	}
	mac := hmac.New(sha256.New, []byte(hk))

	if len(content) < mac.Size() {
		return models.Session{}, fmt.Errorf("invalid sessionid")
	}

	macBytes := content[:mac.Size()]
	messageBytes := content[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return models.Session{}, fmt.Errorf("invalid sessionid")
	}

	if err := json.Unmarshal(messageBytes, &session); err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (hk hmacKey) packUploadState(lus models.Session) (string, error) {
	mac := hmac.New(sha256.New, []byte(hk))
	p, err := json.Marshal(lus)
	if err != nil {
		return "", err
	}

	mac.Write(p)
	return base64.URLEncoding.EncodeToString(append(mac.Sum(nil), p...)), nil
}
