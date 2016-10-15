package module

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/signature"
)

var FileLock sync.RWMutex
var ManiLock sync.RWMutex

var Apis = []string{"images", "tarsum"}

func GetImagePath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%s/%s/%s", setting.DockyardPath, Apis[apiversion], imageId)
}

func GetLayerPath(imageId, layerfile string, apiversion int64) string {
	return fmt.Sprintf("%s/%s/%s/%s", setting.DockyardPath, Apis[apiversion], imageId, layerfile)
}

// NewURLFromRequest uses information from an *http.Request to
// construct the url.
func NewURLFromRequest(r *http.Request) *url.URL {
	var scheme string

	forwardedProto := r.Header.Get("X-Forwarded-Proto")
	schemeHeader := r.Header.Get("Scheme")
	switch {
	case len(schemeHeader) > 0:
		scheme = schemeHeader
	case len(forwardedProto) > 0:
		scheme = forwardedProto
	case r.TLS != nil:
		scheme = "https"
	case len(r.URL.Scheme) > 0:
		scheme = r.URL.Scheme
	default:
		scheme = "http"
	}

	host := r.Host
	forwardedHost := r.Header.Get("X-Forwarded-Host")
	if len(forwardedHost) > 0 {
		hosts := strings.SplitN(forwardedHost, ",", 2)
		host = strings.TrimSpace(hosts[0])
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
	}

	return u
}

func parseIP(ipStr string) net.IP {
	ip := net.ParseIP(ipStr)

	return ip
}

// RemoteAddr extracts the remote address of the request, taking into
// account proxy headers.
func RemoteAddr(r *http.Request) string {
	// X-Forwarded-For's format: client1, proxy1, proxy2
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		proxies := strings.Split(prior, ",")
		if len(proxies) > 0 {
			remoteAddr := strings.Trim(proxies[0], " ")
			if parseIP(remoteAddr) != nil {
				return remoteAddr
			}
		}
	}
	// X-Real-Ip is less supported, but worth checking in the
	// absence of X-Forwarded-For
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		if parseIP(realIP) != nil {
			return realIP
		}
	}

	return r.RemoteAddr
}

// RemoteIP extracts the remote IP of the request, taking into
// account proxy headers.
func RemoteIP(r *http.Request) string {
	addr := RemoteAddr(r)

	// Try parsing it as "IP:port"
	if ip, _, err := net.SplitHostPort(addr); err == nil {
		return ip
	}

	return addr
}

func ParseManifest(namespace, repository, tag, agent string, data []byte) (error, int64) {
	var oldsumlist []string
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err, 0
	}

	// To check blobs of manifest already uploaded or not
	tarsumlist, err := GetTarsumlist(data)
	if err != nil {
		return err, 0
	}
	if len(tarsumlist) <= 0 {
		return fmt.Errorf("Manifest content error: no blobs"), 0
	}

	//recycle image
	isok := true
	defer func() {
		if isok == false {
			Recycleimage(tarsumlist)
		}
	}()

	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.IsExist(); err != nil {
			isok = false
			return err, 0
		} else if !available {
			isok = false
			return fmt.Errorf("Manifest content error: blob of manifest is not existed: %s", tarsum), 0
		}
	}
	tagexist := false
	schemaVersion := int64(manifest["schemaVersion"].(float64))
	if schemaVersion == 1 {
		name := namespace + "/" + repository
		if strings.Compare(name, manifest["name"].(string)) != 0 {
			isok = false
			return fmt.Errorf("Manifest content error: namespace and repository is not equel manifest name %s:%s",
				name, manifest["name"].(string)), 0
		}
		if strings.Compare(tag, manifest["tag"].(string)) != 0 {
			isok = false
			return fmt.Errorf("Manifest content error: tag is not equel manifest tag %s:%s", tag, manifest["tag"].(string)), 0
		}

		r := new(models.DockerV2)
		r.Namespace, r.Repository = namespace, repository
		condition := new(models.DockerV2)
		*condition = *r
		r.SchemaVersion, r.Agent = "DOCKERAPIV2", agent
		if err := r.Save(condition); err != nil {
			isok = false
			return fmt.Errorf("Failed to save repository %s/%s :%s", namespace, repository, err.Error()), 0
		}

		//To get the existence of tag; if tag exists, the reference of blob will be updated
		//if not, the reference of blob will be added
		oldtag := new(models.DockerTagV2)
		oldtag.DockerV2, oldtag.Tag = r.Id, tag
		if exits, err := oldtag.IsExist(); err != nil {
			isok = false
			return err, 0
		} else if exits {
			tagexist = true
			old := []byte(oldtag.Manifest)
			if oldsumlist, err = GetTarsumlist(old); err != nil {
				isok = false
				return err, 0
			}
		}
		for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
			v := manifest["history"].([]interface{})[k]
			compatibility := v.(map[string]interface{})["v1Compatibility"].(string)

			var image map[string]interface{}
			if err := json.Unmarshal([]byte(compatibility), &image); err != nil {
				isok = false
				return err, 0
			}

			if k == 0 {
				t := new(models.DockerTagV2)
				t.DockerV2, t.Tag = r.Id, tag
				condition := new(models.DockerTagV2)
				*condition = *t
				t.ImageId, t.Manifest, t.Schema = image["id"].(string), string(data), schemaVersion
				if err := t.Save(condition); err != nil {
					isok = false
					// TODO: recycle tag table
					if err := recycletag(namespace, repository, tag); err != nil {
						return err, 0
					}
					return err, 0
				}

			}
		}
	} else if schemaVersion == 2 {
		r := new(models.DockerV2)
		r.Namespace, r.Repository = namespace, repository
		condition := new(models.DockerV2)
		*condition = *r
		r.SchemaVersion, r.Agent = "DOCKERAPIV2", agent
		if err := r.Save(condition); err != nil {
			isok = false
			return fmt.Errorf("Failed to save repository %s/%s :%s", namespace, repository, err.Error()), 0
		}
		//To get the existence of tag; if tag exists, the reference of blob will be updated
		//if not, the reference of blob will be added
		oldtag := new(models.DockerTagV2)
		oldtag.DockerV2, oldtag.Tag = r.Id, tag
		if exits, err := oldtag.IsExist(); err != nil {
			isok = false
			return err, 0
		} else if exits {
			tagexist = true
			old := []byte(oldtag.Manifest)
			if oldsumlist, err = GetTarsumlist(old); err != nil {
				isok = false
				return err, 0
			}
		}

		confblobsum := manifest["config"].(map[string]interface{})["digest"].(string)
		imageId := strings.Split(confblobsum, ":")[1]
		t := new(models.DockerTagV2)
		t.DockerV2, t.Tag = r.Id, tag
		cond := new(models.DockerTagV2)
		*cond = *t
		t.ImageId, t.Manifest, t.Schema = imageId, string(data), schemaVersion
		if err := t.Save(cond); err != nil {
			isok = false
			// TODO: recycle tag table
			if err := recycletag(namespace, repository, tag); err != nil {
				return err, 0
			}
			return err, 0
		}

	} else {
		isok = false
		return fmt.Errorf("Manifest content error: invalid schema version"), 0
	}

	if err := UpdateImgRefCnt(tarsumlist); err != nil {
		// TODO: recycle tag table and images
		if err := recycletag(namespace, repository, tag); err != nil {
			return err, 0
		}
		return err, 0
	}

	if len(oldsumlist) > 0 {
		/*
			if err := DeleteImgRefCnt(oldsumlist); err != nil {
				// TODO: recycle tag table and images
				return err, 0
			}
		*/
		DeleteImgRefCnt(tarsumlist, oldsumlist, tagexist)
	}

	return nil, schemaVersion
}

func recycletag(namespace, repository, tag string) error {
	r := new(models.DockerV2)
	r.Namespace, r.Repository = namespace, repository
	if exist, err := r.IsExist(); err != nil {
		return fmt.Errorf("recycle tag error: %s", err)
	} else if exist {
		return fmt.Errorf("recycle tag error: repository is not exist")
	}

	t := new(models.DockerTagV2)
	t.DockerV2, t.Tag = r.Id, tag
	if exists, err := t.IsExist(); err != nil {
		return fmt.Errorf("recycle tag error: %s", err)
	} else if exists {
		if err := t.Delete(); err != nil {
			return fmt.Errorf("recycle tag error: %s", err)
		}
	}
	return nil
}

func Recycleimage(tarsumlist []string) error {
	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.IsExist(); err != nil {
			return fmt.Errorf("recycle image error: %s", err)
		} else if !available {
			continue
		}

		if i.Reference == 0 {
			if err := i.Delete(); err != nil {
				return fmt.Errorf("recycle image error: %s", err)
			}
			imagePath := GetImagePath(tarsum, setting.DOCKERAPIV2)
			if err := RemoveFlieWithLock(imagePath); err != nil {
				return fmt.Errorf("recycle image error: %s", err)
			}
		}
	}
	return nil
}

func recycleupdatedimage(tarsumlist []string) error {
	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.IsExist(); err != nil {
			return fmt.Errorf("recycle updated image error: %s", err)
		} else if !available {
			return fmt.Errorf("updated image is not exist")
		}

		if i.Reference == 0 {
			return fmt.Errorf("updated image reference is zero")
		} else {
			i.Reference = i.Reference - 1
			if i.Reference == 0 {
				if err := i.Delete(); err != nil {
					return fmt.Errorf("recycle updated image error: %s", err)
				}
				imagePath := GetImagePath(tarsum, setting.DOCKERAPIV2)
				if err := RemoveFlieWithLock(imagePath); err != nil {
					return fmt.Errorf("recycle updated image error: %s", err)
				}
			} else {
				if err := i.Update(); err != nil {
					return fmt.Errorf("recycle updated image error: %s", err)
				}
			}
		}
	}
	return nil
}

func GetTarsumlist(data []byte) ([]string, error) {
	var tarsumlist []string
	var layers = []string{"", "fsLayers", "layers"}
	var tarsums = []string{"", "blobSum", "digest"}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return []string{}, err
	}

	schemaVersion := int64(manifest["schemaVersion"].(float64))
	if schemaVersion == 2 {
		confblobsum := manifest["config"].(map[string]interface{})["digest"].(string)
		tarsum := strings.Split(confblobsum, ":")[1]
		tarsumlist = append(tarsumlist, tarsum)
	}

	section := layers[schemaVersion]
	item := tarsums[schemaVersion]
	for i := len(manifest[section].([]interface{})) - 1; i >= 0; i-- {
		blobsum := manifest[section].([]interface{})[i].(map[string]interface{})[item].(string)
		tarsum := strings.Split(blobsum, ":")[1]
		tarsumlist = append(tarsumlist, tarsum)
	}

	return tarsumlist, nil
}

//image reference counting increased when repository upload successfully
func UpdateImgRefCnt(tarsumlist []string) error {
	if len(tarsumlist) <= 0 {
		return fmt.Errorf("no blobs")
	}

	recycletarsum := []string{}
	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.IsExist(); err != nil {
			if err := recycleupdatedimage(recycletarsum); err != nil {
				return err
			}
			return err
		} else if !available {
			return fmt.Errorf("blobs not existed")
		}

		i.Reference = i.Reference + 1
		if err := i.Update(); err != nil {
			if err := recycleupdatedimage(recycletarsum); err != nil {
				return err
			}
			return err

		}
		recycletarsum = append(recycletarsum, tarsum)
	}

	return nil
}

//old image reference counting decreased when repository upload successfully
func DeleteImgRefCnt(tarsumlist, oldsumlist []string, tagexist bool) error {
	if len(oldsumlist) == 0 {
		return nil
	}
	for _, oldtarsum := range oldsumlist {
		innew := false
		for _, tarsum := range tarsumlist {
			if strings.Compare(tarsum, oldtarsum) == 0 {
				innew = true
				break
			}
		}

		i := new(models.DockerImageV2)
		i.BlobSum = oldtarsum
		if available, err := i.IsExist(); err != nil {
			return err
		} else if !available {
			return fmt.Errorf("blobs not existed")
		}

		i.Reference = i.Reference - 1
		if i.Reference == 0 {
			if innew {
				continue
			}
			if err := i.Delete(); err != nil {
				return err
			}
			imagePath := GetImagePath(oldtarsum, setting.DOCKERAPIV2)
			if err := RemoveFlieWithLock(imagePath); err != nil {
				return err
			}
		} else {
			if err := i.Update(); err != nil {
				return err
			}
		}
	}

	return nil
}

//if digest of tag accord with the reference, then delete the tag info
func DeleteTagByRefer(id int64, reference string, tagslist []string) error {
	tagexists := false
	for _, tag := range tagslist {
		t := new(models.DockerTagV2)
		t.DockerV2, t.Tag = id, tag
		if exists, err := t.IsExist(); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("Tag is not exists!")
		}

		digest, err := signature.DigestManifest([]byte(t.Manifest))
		if err != nil {
			return err
		}

		if strings.Compare(digest, reference) == 0 {
			tagexists = true

			/*			if tarsumlist, err := GetTarsumlist([]byte(t.Manifest)); err != nil {
							return err
						} else {
							for _, tarsum := range tarsumlist {
								i := new(models.DockerImageV2)
								i.BlobSum = tarsum

								if available, err := i.IsExist(); err != nil {
									return err
								} else if !available {
									return fmt.Errorf("Blob is not exists!")
								}

								i.Reference = i.Reference - int64(1)
								if err := i.Update(); err != nil {
									return err
								}
							}
						}
			*/
			if err := t.Delete(); err != nil {
				return err
			}
		}
	}
	if tagexists == false {
		return fmt.Errorf("Tag is not exists!")
	}

	/*
		if err := UpdateDockerV2(id); err != nil {
			return err
		}
	*/

	return nil
}

func Deleteblobv2(tarsum string) error {
	i := new(models.DockerImageV2)
	i.BlobSum = tarsum
	if available, err := i.IsExist(); err != nil {
		return err
	} else if !available {
		return fmt.Errorf("Not found docker blob")
	}

	if i.Reference == 0 {
		if err := i.Delete(); err != nil {
			return err
		}
		imagePath := GetImagePath(tarsum, setting.DOCKERAPIV2)
		if err := RemoveFlieWithLock(imagePath); err != nil {
			return err
		}
	} else if i.Reference > 0 {
		if i.Reference = i.Reference - int64(1); i.Reference == 0 {
			if err := i.Delete(); err != nil {
				return err
			}
			imagePath := GetImagePath(tarsum, setting.DOCKERAPIV2)
			if err := RemoveFlieWithLock(imagePath); err != nil {
				return err
			}
		} else {
			if err := i.Update(); err != nil {
				return err
			}
		}
	} else {

	}

	return nil
}

//Update repository info in db
func UpdateDockerV2(id int64) error {
	t := new(models.DockerTagV2)
	t.DockerV2 = id
	if exists, err := t.IsExist(); err != nil {
		return err
	} else if !exists {
		r := new(models.DockerV2)
		r.Id = id
		if err := r.Delete(); err != nil {
			return err
		}
	}

	return nil
}

func SaveLayerLocal(srcPath, srcFile, dstPath, dstFile string, reqbody io.Reader) (int64, error) {
	digest := path.Base(dstPath)
	if !utils.IsDirExist(dstPath) {
		os.MkdirAll(dstPath, 0750)
	}

	var layerlen int64

	FileLock.Lock()
	defer FileLock.Unlock()
	if fifo, err := os.Stat(srcFile); err == nil {
		//docker 1.9.x above version saves layer in PATCH methord
		file, err := os.Open(srcFile)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		sha256h := sha256.New()
		if _, err := io.Copy(sha256h, file); err != nil {
			return 0, fmt.Errorf("Generate data hash code error %s", err.Error())
		}
		hash256 := fmt.Sprintf("%x", sha256h.Sum(nil))
		if isEqual := strings.Compare(digest, hash256); isEqual != 0 {
			os.RemoveAll(srcPath)
			return 0, fmt.Errorf("App hash is not equel digest %s:%s", hash256, digest)
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			return 0, err
		}
		os.RemoveAll(srcPath)
		layerlen = fifo.Size()
	} else {
		//docker 1.9.x below version saves layer in PUT methord
		file, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		size, err := io.Copy(file, reqbody)
		if err != nil {
			return 0, err
		}
		layerlen = size

		sha256h := sha256.New()
		if _, err := file.Seek(0, 0); err != nil {
			return 0, err
		}
		if _, err := io.Copy(sha256h, file); err != nil {
			return 0, fmt.Errorf("Generate data hash code error %s", err.Error())
		}
		hash256 := fmt.Sprintf("%x", sha256h.Sum(nil))
		if isEqual := strings.Compare(digest, hash256); isEqual != 0 {
			os.RemoveAll(dstPath)
			return 0, fmt.Errorf("App hash is not equel digest %s:%s", hash256, digest)
		}
	}

	return layerlen, nil
}

func RemoveFlieWithLock(path string) error {
	FileLock.Lock()
	defer FileLock.Unlock()
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}
