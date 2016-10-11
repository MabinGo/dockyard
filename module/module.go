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
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/containerops/dockyard/db"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils/uuid"
	"github.com/containerops/dockyard/utils/validate"
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
	INVALID_PARAM         = "INVALID_PARAM"
	MISSING_PARAM         = "MISSING_PARAM"
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
	return fmt.Sprintf("%s/%s/%s", setting.DockyardPath, Apis[apiversion], imageId)
}

func GetManifestPath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%s/%s/%s/manifest", setting.DockyardPath, Apis[apiversion], imageId)
}

func GetSignaturePath(imageId, signfile string, apiversion int64) string {
	return fmt.Sprintf("%s/%s/%s/%s", setting.DockyardPath, Apis[apiversion], imageId, signfile)
}

func GetLayerPath(imageId, layerfile string, apiversion int64) string {
	return fmt.Sprintf("%s/%s/%s/%s", setting.DockyardPath, Apis[apiversion], imageId, layerfile)
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

const (
	PUSH   = "push"
	PULL   = "pull"
	DELETE = "delete"
)

const (
	RUNNING = "running"
	END     = "end"
)

var sLock sync.RWMutex

/*
func sessionLock() {
	//sLock.Lock()
	s := new(models.Session)
	s.TableLock()
}

func sessionUnlock() {
	s := new(models.Session)
	s.TableUnlock()
	//sLock.Unlock()
}
*/
func SessionLock(namespace, repository, action string, version int64) error {
	sessionTableLock()
	defer sessionTableUnlock()

	current := new(models.Session)
	exists, err := current.Read(namespace, repository, version)
	if err != nil {
		return err
	}
	switch action {
	case PULL:
		if !exists {
			current.Namespace = namespace
			current.Repository = repository
			current.Version = version
			//current.UUID = uuid
			current.Locked++
			if err := current.Save(namespace, repository, version); err != nil {
				return err
			}
		} else {
			if current.Locked < 0 {
				return fmt.Errorf("%s/%s source is busy", namespace, repository)
			} else {
				current.Locked++
				if err := current.Save(namespace, repository, version); err != nil {
					return err
				}
			}
		}
	case DELETE:
		if !exists {
			current.Namespace = namespace
			current.Repository = repository
			current.Version = version
			//current.UUID = uuid
			current.Locked = -1
			if err := current.Save(namespace, repository, version); err != nil {
				return err
			}
		} else {
			if current.Locked != 0 {
				return fmt.Errorf("%s/%s source is busy", namespace, repository)
			} else {
				current.Locked = -1
				if err := current.Save(namespace, repository, version); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("bad action %s", action)
	}

	return nil
}

func SessionUnlock(namespace, repository, action string, version int64) error {
	sessionTableLock()
	defer sessionTableUnlock()

	current := new(models.Session)
	exists, err := current.Read(namespace, repository, version)
	if err != nil {
		return err
	} else if !exists {
		return nil
	}

	switch action {
	case PULL:
		if current.Locked > 0 {
			current.Locked--
		} else if current.Locked < 0 {
			return fmt.Errorf("%s/%s bad lock status", namespace, repository)
		} else {

		}
	case DELETE:
		current.Locked = 0
	default:
		return fmt.Errorf("bad action %s", action)
	}

	if current.Locked == 0 {
		return current.Delete()
	}

	if err := current.UpdateSessionLock(current.Locked); err != nil {
		return err
	}

	return nil
}

// push session
func GenerateSessionID(namespace, repository string, version int64) (string, string, error) {
	sessionTableLock()
	defer sessionTableUnlock()

	hk := hmacKey(DYSESSIONID)
	origin := new(models.Session)
	exists, err := origin.Read(namespace, repository, version)
	if err != nil {
		return "", "", err
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", "", err
	}

	current := new(models.Session)
	//if namespace/repository is existed, it means concurrent
	if exists {
		//if origin.Locked == -1) || (origin.Locked > 0) {
		if origin.Locked != 0 {
			return "", "", fmt.Errorf("%s/%s source is busy", namespace, repository)
		} else {
			origin.Locked = -1
			origin.UUID = uuid
			origin.Version = version
			*current = *origin
		}
	} else { //if it is not existed, new create
		current.Namespace = namespace
		current.Repository = repository
		current.Version = version
		current.UUID = uuid
		current.Locked = -1
	}

	sessionid, err := hk.packUploadSession(*current)
	if err != nil {
		return "", "", err
	}

	if err := current.Save(namespace, repository, version); err != nil {
		return "", "", err
	}

	return uuid, sessionid, nil
}

// push session
func ValidateSessionID(namespace, repository, sessionid string, version int64) (models.Session, error) {
	hk := hmacKey(DYSESSIONID)
	s, err := hk.unpackUploadSession(sessionid)
	if err != nil {
		return models.Session{}, err
	}

	if (s.Namespace != namespace) || (s.Repository != repository) {
		return models.Session{}, fmt.Errorf("bad App-Upload-UUID, mismatch repository %s/%s", namespace, repository)
	}

	//sessionLock()
	//defer sessionUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return models.Session{}, err
	} else if !exists {
		return models.Session{}, fmt.Errorf("not found repository %s/%s", namespace, repository)
	}

	// TODO: identify the same session
	if (current.Locked != s.Locked) || (current.UUID != s.UUID) {
		return models.Session{}, fmt.Errorf("%s/%s source is busy", namespace, repository)
	}

	return s, nil
}

// push session
func ReleaseSessionID(namespace, repository string, version int64, state string) error {
	//sessionLock()
	//defer sessionUnlock()

	sessionTableLock()
	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		sessionTableUnlock()
		return err
	} else if !exists {
		sessionTableUnlock()
		return nil
	} else {
		sessionTableUnlock()
		if state == END {
			artifactTableLock()
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			exists, err := i.IsExist()
			if err != nil {
				artifactTableUnlock()
				return err
			} else if !exists {
				artifactTableUnlock()
				return fmt.Errorf("not found blob %d", i.Id)
			}

			i.Active = 1
			if err := i.UpdateImageStatus(i.Active); err != nil {
				artifactTableUnlock()
				return err
			}
		}

		if state != END && current.Locked == -1 {
			// delete images table and blob
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			if blobsum, _ := i.Delete(); blobsum != "" {
				os.RemoveAll(fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", blobsum))
			}
		}
		artifactTableUnlock()

		sessionTableLock()
		if err := current.Delete(); err != nil {
			sessionTableUnlock()
			return err
		}
		sessionTableUnlock()

		return nil
	}
}

func SaveImageID(namespace, repository string, imgid, version int64) error {
	sessionTableLock()
	defer sessionTableUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found %s/%s", namespace, repository)
	}

	current.Imageid = imgid

	if err := current.Save(namespace, repository, version); err != nil {
		return err
	}

	return nil
}

// recycle push session
func RecycleSession(namespace, repository string, version int64) (bool, error) {
	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return false, err
	} else if !exists {
		return false, nil
	} else {
		sessionTableLock()
		if current.Locked == -1 {
			// delete images table and blob
			artifactTableLock()
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			if blobsum, _ := i.Delete(); blobsum != "" {
				os.RemoveAll(fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", blobsum))
			}
			artifactTableUnlock()
		}

		if err := current.Delete(); err != nil {
			sessionTableUnlock()
			return true, err
		}

		sessionTableUnlock()
		return true, nil
	}
}

func RecycleSourceThread() {
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(setting.RecycleInterval))
			se := new(models.Session)
			records := []models.Session{}
			if cnt, err := se.Find(&records); err != nil {
				log.Errorf("recycle from db error : %s", err.Error())
			} else if cnt == 0 {
				continue
			}

			for _, v := range records {
				if time.Since(v.UpdatedAt).Seconds() >= float64(setting.RecycleInterval) {
					s := new(models.Session)
					*s = v

					sessionTableLock()
					if s.Locked <= 0 {
						sessionTableUnlock()

						artifactTableLock()
						// delete images table and blob
						i := new(models.ArtifactV1)
						i.Id = v.Imageid

						if blobsum, err := i.Delete(); err != nil {
							log.Errorf("recycle source artifact error : %s", err.Error())
						} else if blobsum != "" {
							os.RemoveAll(fmt.Sprintf("%s/%s/%s", setting.DockyardPath, "app", blobsum))
						}
						artifactTableUnlock()
					}

					sessionTableLock()
					if err := s.Delete(); err != nil {
						log.Errorf("recycle source session error : %s", err.Error())
					}
					sessionTableUnlock()
				}
			}
		}
	}()
}

func ValidateName(namespace, repository string) (string, error) {
	if !validate.IsNameValid(namespace) {
		return NAME_INVALID, fmt.Errorf("Invalid namespace format : %s", namespace)
	}

	if !validate.IsRepoValid(repository) {
		return NAME_INVALID, fmt.Errorf("Invalid repository format : %s", repository)
	}

	return "", nil
}

func ValidateParams(system, arch, appname, tag string) (string, error) {
	osLen := len(system)
	if osLen == 0 || osLen > 128 {
		return NAME_INVALID, fmt.Errorf("Invalid os name length")
	}

	archLen := len(arch)
	if archLen == 0 || archLen > 128 {
		return NAME_INVALID, fmt.Errorf("Invalid arch name length")
	}

	if !validate.IsAppValid(appname) {
		return NAME_INVALID, fmt.Errorf("Invalid app format: %s", appname)
	}

	if !validate.IsTagValid(tag) {
		return TAG_INVALID, fmt.Errorf("Invalid tag format: %s", tag)
	}

	return "", nil
}

func ValidateDigest(digest string) (string, error) {
	if !validate.IsDigestValid(digest) {
		return DIGEST_INVALID, fmt.Errorf("Invalid digest format: %s", digest)
	}

	return "", nil
}

func ValidateTag(tag string) (string, error) {
	if !validate.IsTagValid(tag) {
		return TAG_INVALID, fmt.Errorf("Invalid tag format: %s", tag)
	}

	return "", nil
}

//Table lock
func sessionTableLock() error {
	cmd := "SET AUTOCOMMIT=0"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = fmt.Sprintf("LOCK TABLES sessions WRITE")
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func sessionTableUnlock() error {
	cmd := "COMMIT"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = "UNLOCK TABLES"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func artifactTableLock() error {
	cmd := "SET AUTOCOMMIT=0"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = fmt.Sprintf("LOCK TABLES artifact_v1 WRITE")
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	return nil
}

func artifactTableUnlock() error {
	cmd := "COMMIT"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = "UNLOCK TABLES"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	return nil
}

//Row lock
func sessionRowLock(s models.Session) error {
	cmd := "SET AUTOCOMMIT=0"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = fmt.Sprintf("select * from sessions where namespace=%s and repository=%s and version=%d for update", s.Namespace, s.Repository, s.Version)
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func sessionRowUnlock() error {
	cmd := "COMMIT"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	return nil
}

func artifactRowLock(a models.ArtifactV1) error {
	cmd := "SET AUTOCOMMIT=0"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = fmt.Sprintf("select * from artifact_v1 where app_v1=%d and os=%s and arch=%s and app=%s and tag=%s for update",
		a.AppV1, a.OS, a.Arch, a.App, a.Tag)
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func artifactRowUnlock() error {
	cmd := "COMMIT"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	return nil
}
