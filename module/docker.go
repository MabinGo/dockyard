package module

import (
	"encoding/json"
	"fmt"
	"io"
	//"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/signature"
)

// NewURLFromRequest uses information from an *http.Request to
// construct the url.
func NewURLFromRequest(r *http.Request) *url.URL {
	var scheme string

	forwardedProto := r.Header.Get("X-Forwarded-Proto")

	switch {
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

func ParseManifest(id int64, tag string, data []byte) (error, int64) {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err, 0
	}

	//To get the existence of tag; if tag exists, the reference of blob will be updated
	//if not, the reference of blob will be added
	var oldsumlist []string
	oldtag := new(models.DockerTagV2)
	oldtag.DockerV2, oldtag.Tag = id, tag
	if exits, err := oldtag.IsExist(); err != nil {
		return err, 0
	} else if exits {
		old := []byte(oldtag.Manifest)
		if oldsumlist, err = GetTarsumlist(old); err != nil {
			return err, 0
		}
	}

	schemaVersion := int64(manifest["schemaVersion"].(float64))
	if schemaVersion == 1 {
		for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
			v := manifest["history"].([]interface{})[k]
			compatibility := v.(map[string]interface{})["v1Compatibility"].(string)

			var image map[string]interface{}
			if err := json.Unmarshal([]byte(compatibility), &image); err != nil {
				return err, 0
			}

			if k == 0 {
				t := new(models.DockerTagV2)
				t.DockerV2, t.Tag = id, tag
				condition := new(models.DockerTagV2)
				*condition = *t
				t.ImageId, t.Manifest, t.Schema = image["id"].(string), string(data), schemaVersion
				if err := t.Save(condition); err != nil {
					return err, 0
				}

			}
		}
	} else if schemaVersion == 2 {
		t := new(models.DockerTagV2)
		t.DockerV2, t.Tag = id, tag
		condition := new(models.DockerTagV2)
		*condition = *t
		t.ImageId, t.Manifest, t.Schema = "schemaV2", string(data), schemaVersion
		if err := t.Save(condition); err != nil {
			return err, 0
		}

	} else {
		return fmt.Errorf("invalid schema version"), 0
	}

	if tarsumlist, err := GetTarsumlist(data); err != nil {
		return err, 0
	} else {
		if err := UpdateImgRefCnt(tarsumlist); err != nil {
			return err, 0
		}
	}
	if len(oldsumlist) > 0 {
		if err := DeleteImgRefCnt(oldsumlist); err != nil {
			return err, 0
		}
	}

	return nil, schemaVersion
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
	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.Write(); err != nil {
			return err
		} else if !available {
			return fmt.Errorf("blobs not existed")
		}

		i.Reference = i.Reference + 1
		if err := i.Update(); err != nil {
			return err
		}
	}

	return nil
}

//old image reference counting decreased when repository upload successfully
func DeleteImgRefCnt(tarsumlist []string) error {
	if len(tarsumlist) <= 0 {
		return fmt.Errorf("no blobs")
	}
	for _, tarsum := range tarsumlist {
		i := new(models.DockerImageV2)
		i.BlobSum = tarsum
		if available, err := i.Write(); err != nil {
			return err
		} else if !available {
			return fmt.Errorf("blobs not existed")
		}

		i.Reference = i.Reference - 1
		if i.Reference == 0 {
			if err := i.Delete(); err != nil {
				return err
			}
			imagePath := GetImagePath(tarsum, setting.DOCKERAPIV2)
			if err := os.RemoveAll(imagePath); err != nil {
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

			if tarsumlist, err := GetTarsumlist([]byte(t.Manifest)); err != nil {
				return err
			} else {
				for _, tarsum := range tarsumlist {
					i := new(models.DockerImageV2)
					i.BlobSum = tarsum

					if available, err := i.Write(); err != nil {
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

			if err := t.Delete(); err != nil {
				return err
			}
		}
	}
	if tagexists == false {
		return fmt.Errorf("Tag is not exists!")
	}

	if err := UpdateDockerV2(id); err != nil {
		return err
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
	if !utils.IsDirExist(dstPath) {
		os.MkdirAll(dstPath, os.ModePerm)
	}

	if utils.IsFileExist(dstFile) {
		os.Remove(dstFile)
	}

	var layerlen int64
	if fifo, err := os.Stat(srcFile); err == nil {
		//docker 1.9.x above version saves layer in PATCH methord
		if err := os.Rename(srcFile, dstFile); err != nil {
			return 0, err
		}
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
	}

	return layerlen, nil
}
