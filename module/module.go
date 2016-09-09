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

func (hk hmacKey) unpackUploadSession(sessionid string) (models.Session, error) {
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

func (hk hmacKey) packUploadSession(lus models.Session) (string, error) {
	mac := hmac.New(sha256.New, []byte(hk))
	p, err := json.Marshal(lus)
	if err != nil {
		return "", err
	}

	mac.Write(p)
	return base64.URLEncoding.EncodeToString(append(mac.Sum(nil), p...)), nil
}

type hmacKey string

var DYSESSIONID = "dockyardsessionid"

var StateLock sync.RWMutex

func sessionLock() {
	StateLock.Lock()
}

func sessionUnlock() {
	StateLock.Unlock()
}

func GetAction(method string) string {
	switch method {
	case "POST", "PUT", "PATCH":
		return "push"
	case "HEAD", "GET":
		return "pull"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

func SessionLock(namespace, repository, action string, version int64) error {
	sessionLock()
	defer sessionUnlock()
	/*
		uuid, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
	*/
	current := new(models.Session)
	exists, err := current.Read(namespace, repository, version)
	if err != nil {
		return err
	}
	switch action {
	case "pull":
		if !exists {
			current.Namespace = namespace
			current.Repository = repository
			//current.UUID = uuid
			current.Locked++
			if err := current.Save(namespace, repository, version); err != nil {
				// TODO: what would be handled when failure
				return err
			}
		} else {
			if current.Locked < 0 {
				return fmt.Errorf("%v/%v is busy", namespace, repository)
			} else {
				current.Locked++
				if err := current.Save(namespace, repository, version); err != nil {
					// TODO: what would be handled when failure
					return err
				}
			}
		}
	case "delete":
		if !exists {
			current.Namespace = namespace
			current.Repository = repository
			//current.UUID = uuid
			current.Locked = -1
			if err := current.Save(namespace, repository, version); err != nil {
				// TODO: what would be handled when failure
				return err
			}
		} else {
			if current.Locked != 0 {
				return fmt.Errorf("%v/%v is busy", namespace, repository)
			} else {
				current.Locked = -1
				if err := current.Save(namespace, repository, version); err != nil {
					// TODO: what would be handled when failure
					return err
				}
			}
		}
	default:
		return fmt.Errorf("bad action %v", action)
	}

	return nil
}

func SessionUnlock(namespace, repository, action string, version int64) error {
	sessionLock()
	defer sessionUnlock()

	current := new(models.Session)
	exists, err := current.Read(namespace, repository, version)
	if err != nil {
		return err
	}

	if exists {
		switch action {
		case "pull":
			if current.Locked > 0 {
				current.Locked--
				if err := current.Save(namespace, repository, version); err != nil {
					// TODO: what would be handled when failure
					return err
				}
			} else {
				return fmt.Errorf("%v/%v is busy", namespace, repository)
			}
		case "delete":
			current.Locked = 0
			if err := current.Save(namespace, repository, version); err != nil {
				// TODO: what would be handled when failure
				return err
			}
		default:
			return fmt.Errorf("bad action %v", action)
		}
	}

	return nil
}

// push session
func GenerateSessionID(namespace, repository string, version int64) (string, error) {
	sessionLock()
	defer sessionUnlock()

	hk := hmacKey(DYSESSIONID)
	origin := new(models.Session)
	exists, err := origin.Read(namespace, repository, version)
	if err != nil {
		return "", err
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	current := new(models.Session)
	//if namespace/repository is existed, it means concurrent
	if exists {
		if (origin.Locked == -1) || (origin.Locked > 0) {
			return "", fmt.Errorf("%v/%v is busy", namespace, repository)
		} else {
			origin.Locked = -1
			origin.UUID = uuid
			*current = *origin
		}
	} else { //if it is not existed, new create
		current.Namespace = namespace
		current.Repository = repository
		current.UUID = uuid
		current.Locked = -1
	}

	sessionid, err := hk.packUploadSession(*current)
	if err != nil {
		return "", err
	}
	fmt.Printf("\n #### mabin GetSessionID 000: %v \n", err)

	if err := current.Save(namespace, repository, version); err != nil {
		return "", err
	}

	return sessionid, nil
}

// push session
func ValidateSessionID(namespace, repository, sessionid string, version int64) error {
	hk := hmacKey(DYSESSIONID)
	s, err := hk.unpackUploadSession(sessionid)
	if err != nil {
		return err
	}

	if (s.Namespace != namespace) || (s.Repository != repository) {
		return fmt.Errorf("bad App-Upload-UUID, mismatch repository %v/%v", namespace, repository)
	}

	//SessionLock()
	//defer SessionUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found repository %v/%v", namespace, repository)
	}

	// TODO: identify the same session
	if (current.Locked != s.Locked) || (current.UUID != s.UUID) {
		return fmt.Errorf("%v/%v is busy", namespace, repository)
	}
	/*
		if (current.Locked == -1) || (current.Locked > 0) {
			return fmt.Errorf("%v/%v is busy", namespace, repository)
		} else if current.Locked == 0 {
			// TODO:
			if err := current.UpdateLockState(-1); err != nil {
				// TODO: what would be handled when failure
				return err
			}
		}
	*/
	return nil
}

// push session
func ReleaseSessionID(namespace, repository string, version int64) error {
	sessionLock()
	defer sessionUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found repository %v/%v", namespace, repository)
	}

	current.Locked = 0
	if err := current.UpdateSessionLock(current.Locked); err != nil {
		// TODO: lock should be recycled when release failure
		return err
	}

	// TODO: delete session table
	// ...

	return nil
}
