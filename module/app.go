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

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils/uuid"
)

var AppFileLock sync.RWMutex

//create a directory to save image,format is .../$dockyardpath/app/$namespace/$repository/$blobsum
func GetAppImagePath(namespace, repository, blobsum string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", setting.DockyardPath, "app", namespace, repository, blobsum)
}

//save image layer,format is .../$dockyardpath/app/$namespace/$repository/$blobsum/layer
func GetAppLayerPath(namespace, repository, blobsum, layer string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s", setting.DockyardPath, "app", namespace, repository, blobsum, layer)
}

func GetOriginBlobsum(id int64, system, arch, appname, tag string) (blobSum string) {
	o := new(models.ArtifactV1)
	o.AppV1, o.OS, o.Arch, o.App, o.Tag = id, system, arch, appname, tag
	if exists, err := o.Read(); err != nil {
		blobSum = ""
	} else if !exists {
		blobSum = ""
	} else {
		blobSum = o.BlobSum
	}
	return
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

func sessionLock() {
	sLock.Lock()
	//s := new(models.Session)
	//s.TableLock()
}

func sessionUnlock() {
	//s := new(models.Session)
	//s.TableUnlock()
	sLock.Unlock()
}

func SessionLock(namespace, repository, action string, version int64) error {
	sessionLock()
	defer sessionUnlock()

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
	sessionLock()
	defer sessionUnlock()

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
	sessionLock()
	defer sessionUnlock()

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
	sessionLock()
	defer sessionUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return err
	} else if !exists {
		return nil
	} else {
		if state == END {
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			exists, err := i.IsExist()
			if err != nil {
				return err
			} else if !exists {
				return fmt.Errorf("not found blob %d", i.Id)
			}

			i.Active = 1
			if err := i.UpdateImageStatus(i.Active); err != nil {
				return err
			}
		}

		if state != END && current.Locked == -1 {
			// delete images table and blob
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			if blobsum, _ := i.Delete(); blobsum != "" {
				os.RemoveAll(GetAppImagePath(namespace, repository, blobsum))
			}
		}

		if err := current.Delete(); err != nil {
			return err
		}

		return nil
	}
}

func SaveImageID(namespace, repository string, imgid, version int64) error {
	sessionLock()
	defer sessionUnlock()

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
	sessionLock()
	defer sessionUnlock()

	current := new(models.Session)
	if exists, err := current.Read(namespace, repository, version); err != nil {
		return false, err
	} else if !exists {
		return false, nil
	} else {
		if current.Locked == -1 {
			// delete images table and blob
			i := new(models.ArtifactV1)
			i.Id = current.Imageid
			if blobsum, _ := i.Delete(); blobsum != "" {
				os.RemoveAll(GetAppImagePath(namespace, repository, blobsum))
			}
		}

		if err := current.Delete(); err != nil {
			return true, err
		}

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

					sessionLock()
					if s.Locked <= 0 {
						// delete images table and blob
						i := new(models.ArtifactV1)
						i.Id = v.Imageid

						if blobsum, err := i.Delete(); err != nil {
							log.Errorf("recycle source artifact error : %s", err.Error())
						} else if blobsum != "" {
							os.RemoveAll(GetAppImagePath(s.Namespace, s.Repository, blobsum))
						}
					}

					if err := s.Delete(); err != nil {
						log.Errorf("recycle source session error : %s", err.Error())
					}
					sessionUnlock()
				}
			}
		}
	}()
}
