//adapt to docker API
package module

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/libtrust"
	"github.com/gorilla/mux"

	"github.com/containerops/dockyard/backend"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
	"github.com/containerops/dockyard/utils/signature"
)

//adapt to docker errorcode
var Errdesc = make(map[string]string)

var (
	UNKNOWN               = "UNKNOWN"
	DIGEST_INVALID        = "DIGEST_INVALID"
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
)

func init() {
	Errdesc[UNKNOWN] = "unknown error"
	Errdesc[DIGEST_INVALID] = "provided digest did not match uploaded content"
	Errdesc[NAME_INVALID] = "invalid repository name"
	Errdesc[TAG_INVALID] = "manifest tag did not match URI"
	Errdesc[NAME_UNKNOWN] = "repository name not known to registry"
	Errdesc[MANIFEST_UNKNOWN] = "manifest unknown"
	Errdesc[MANIFEST_INVALID] = "manifest invalid"
	Errdesc[MANIFEST_UNVERIFIED] = "manifest failed signature verification"
	Errdesc[MANIFEST_BLOB_UNKNOWN] = "blob unknown to registry"
	Errdesc[BLOB_UNKNOWN] = "blob unknown to registry"
	Errdesc[BLOB_UPLOAD_UNKNOWN] = "blob upload unknown to registry"
	Errdesc[BLOB_UPLOAD_INVALID] = "blob upload invalid"
}

type Errors struct {
	Errors []Errunit `json:"errors"`
}

type Errunit struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

func ReportError(code string, detail interface{}) ([]byte, error) {
	var errs = Errors{}

	item := Errunit{
		Code:    code,
		Message: Errdesc[code],
		Detail:  detail,
	}

	errs.Errors = append(errs.Errors, item)

	return json.Marshal(errs)
}

func ParseManifest(data []byte, namespace, repository, tag string) (error, int64) {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err, 0
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

			i := map[string]string{}
			r := new(models.Repository)

			if k == 0 {
				i["Tag"] = tag
			}
			i["id"] = image["id"].(string)

			if err := r.PutJSONFromManifests(i, namespace, repository, setting.APIVERSION_V2); err != nil {
				return err, 0
			}

			if k == 0 {
				if err := r.PutTagFromManifests(image["id"].(string), namespace, repository, tag, string(data), schemaVersion); err != nil {
					return err, 0
				}
			}
		}
	} else if schemaVersion == 2 {
		r := new(models.Repository)
		if err := r.PutTagFromManifests("schemaV2", namespace, repository, tag, string(data), schemaVersion); err != nil {
			return err, 0
		}
	} else {
		return fmt.Errorf("invalid schema version"), 0
	}

	return nil, schemaVersion
}

//digestSHA256GzippedEmptyTar is the canonical sha256 digest of gzippedEmptyTar
const digestSHA256GzippedEmptyTar = digest.Digest("sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4")

// gzippedEmptyTar is a gzip-compressed version of an empty tar file
var gzippedEmptyTar = []byte{
	31, 139, 8, 0, 0, 9, 110, 136, 0, 255, 98, 24, 5, 163, 96, 20, 140, 88,
	0, 8, 0, 0, 255, 255, 46, 175, 181, 239, 0, 4, 0, 0,
}

//SaveGzippedEmptyTar is to save gzippedEmptyTar image
func SaveGzippedEmptyTar() error {
	tarsum := "a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		return err
	} else if exists {
		return nil
	}
	imagePath := GetImagePath(tarsum, setting.APIVERSION_V2)
	layerPath := GetLayerPath(tarsum, "layer", setting.APIVERSION_V2)

	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}
	if utils.IsFileExist(layerPath) {
		os.Remove(layerPath)
	}

	if err := ioutil.WriteFile(layerPath, gzippedEmptyTar, 0777); err != nil {
		return err
	}

	i.Path, i.Size = layerPath, int64(len(gzippedEmptyTar))
	if err := i.Save(tarsum); err != nil {
		return err
	}

	return nil
}

//SaveV2Conversion is to save schemav2 conversion info
func SaveV2Conversion(namespace, repository, tag string) error {
	t := new(models.Tag)
	if exist, err := t.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("Tag is not exist")
	}

	if t.Schema == 2 {
		var manifest map[string]interface{}
		if err := json.Unmarshal([]byte(t.Manifest), &manifest); err != nil {
			return err
		}
		confblobsum := manifest["config"].(map[string]interface{})["digest"].(string)
		tarsum := strings.Split(confblobsum, ":")[1]

		i := new(models.Image)
		if exist, err := i.Get(tarsum); err != nil {
			return err
		} else if !exist {
			return fmt.Errorf("Config blob is not exist")
		}

		dirPath := path.Dir(i.Path)
		name := path.Base(dirPath)
		fd, err := DownloadLayer(i.Path)
		if err != nil {
			return err
		}
		//fd is schemav2 to schemav1 conversion info
		layer, err := ioutil.ReadAll(fd)
		if err != nil {
			return err
		}
		t.Conversion = string(layer)
		fd.Close()
		os.Remove(GetTmpFile(name))
	}
	if err := t.Save(namespace, repository, tag); err != nil {
		return err
	}

	return nil
}

//ConvertSchema2Sign is to sign signatures
func ConvertSchema2Sign(v1manifest schema1.Manifest) (string, error) {
	p, err := json.MarshalIndent(v1manifest, " ", " ")
	if err != nil {
		return "", err
	}
	trustKey, err := libtrust.GenerateECP256PrivateKey()

	js, err := libtrust.NewJSONSignature(p)
	if err != nil {
		return "", err
	}
	if err := js.Sign(trustKey); err != nil {
		return "", err
	}

	pretty, err := js.PrettySignature("signatures")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

//ConvertSchema2Manifest is schemav2 convert to schemav1
func ConvertSchema2Manifest(t *models.Tag) (string, error) {
	var tarsumlist []string
	var v2manifest map[string]interface{}

	var layers = []string{"", "fsLayers", "layers"}
	var tarsums = []string{"", "blobSum", "digest"}
	if err := json.Unmarshal([]byte(t.Manifest), &v2manifest); err != nil {
		return "", err
	}
	schemaVersion := int64(v2manifest["schemaVersion"].(float64))
	section := layers[schemaVersion]
	item := tarsums[schemaVersion]
	for i := 0; i <= len(v2manifest[section].([]interface{}))-1; i++ {
		blobsum := v2manifest[section].([]interface{})[i].(map[string]interface{})[item].(string)
		tarsumlist = append(tarsumlist, blobsum)
	}

	type imageRootFS struct {
		Type      string          `json:"type"`
		DiffIDs   []digest.Digest `json:"diff_ids,omitempty"`
		BaseLayer string          `json:"base_layer,omitempty"`
	}

	type imageHistory struct {
		Created    time.Time `json:"created"`
		Author     string    `json:"author,omitempty"`
		CreatedBy  string    `json:"created_by,omitempty"`
		Comment    string    `json:"comment,omitempty"`
		EmptyLayer bool      `json:"empty_layer,omitempty"`
	}

	type imageConfig struct {
		RootFS       *imageRootFS   `json:"rootfs,omitempty"`
		History      []imageHistory `json:"history,omitempty"`
		Architecture string         `json:"architecture,omitempty"`
	}

	var img imageConfig
	if err := json.Unmarshal([]byte(t.Conversion), &img); err != nil {
		return "", err
	}

	type v1Compatibility struct {
		ID              string    `json:"id"`
		Parent          string    `json:"parent,omitempty"`
		Comment         string    `json:"comment,omitempty"`
		Created         time.Time `json:"created"`
		ContainerConfig struct {
			Cmd []string
		} `json:"container_config,omitempty"`
		ThrowAway bool `json:"throwaway,omitempty"`
	}
	fsLayerList := make([]schema1.FSLayer, len(img.History))
	history := make([]schema1.History, len(img.History))

	parent := ""
	layerCounter := 0
	var blobsum digest.Digest
	for i, h := range img.History[:len(img.History)-1] {
		if h.EmptyLayer {
			if err := SaveGzippedEmptyTar(); err != nil {
				return "", err
			}
			blobsum = digestSHA256GzippedEmptyTar
		} else {
			if len(img.RootFS.DiffIDs) <= layerCounter {
				return "", fmt.Errorf("too many non-empty layers in History section")
			}
			blobsum = digest.Digest(tarsumlist[layerCounter])
			layerCounter++
		}

		v1ID := digest.FromBytes([]byte(blobsum.Hex() + " " + parent)).Hex()

		if i == 0 && img.RootFS.BaseLayer != "" {
			// windows-only baselayer setup
			baseID := sha512.Sum384([]byte(img.RootFS.BaseLayer))
			parent = fmt.Sprintf("%x", baseID[:32])
		}

		v1Compatibility := v1Compatibility{
			ID:      v1ID,
			Parent:  parent,
			Comment: h.Comment,
			Created: h.Created,
		}
		v1Compatibility.ContainerConfig.Cmd = []string{img.History[i].CreatedBy}
		if h.EmptyLayer {
			v1Compatibility.ThrowAway = true
		}
		jsonBytes, err := json.Marshal(&v1Compatibility)
		if err != nil {
			return "", err
		}

		reversedIndex := len(img.History) - i - 1
		history[reversedIndex].V1Compatibility = string(jsonBytes)
		fsLayerList[reversedIndex] = schema1.FSLayer{BlobSum: blobsum}

		parent = v1ID
	}

	latestHistory := img.History[len(img.History)-1]
	if latestHistory.EmptyLayer {
		if err := SaveGzippedEmptyTar(); err != nil {
			return "", err
		}
		blobsum = digestSHA256GzippedEmptyTar
	} else {
		if len(img.RootFS.DiffIDs) <= layerCounter {
			return "", fmt.Errorf("too many non-empty layers in History section")
		}
		blobsum = digest.Digest(tarsumlist[layerCounter-1])
	}

	fsLayerList[0] = schema1.FSLayer{BlobSum: blobsum}
	dgst := digest.FromBytes([]byte(blobsum.Hex() + " " + parent + " " + t.Conversion))

	// Top-level v1compatibility string should be a modified version of the image config.
	transformedConfig, err := schema1.MakeV1ConfigFromConfig([]byte(t.Conversion), dgst.Hex(), parent, latestHistory.EmptyLayer)
	if err != nil {
		return "", err
	}
	history[0].V1Compatibility = string(transformedConfig)

	v1manifest := schema1.Manifest{
		Versioned: manifest.Versioned{
			SchemaVersion: 1,
		},
		Name:         t.Namespace + "/" + t.Repository,
		Tag:          t.Tag,
		Architecture: img.Architecture,
		FSLayers:     fsLayerList,
		History:      history,
	}

	return ConvertSchema2Sign(v1manifest)
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
func UpdateImgRefCnt(newtarsumlist, oldtarsumlist []string, tagexist bool) error {
	if len(newtarsumlist) <= 0 {
		return fmt.Errorf("no blobs")
	}

	for _, tarsum := range newtarsumlist {
		i := new(models.Image)
		if exists, err := i.Get(tarsum); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("blobs not existed")
		}

		i.Count = i.Count + 1
		if err := i.Save(tarsum); err != nil {
			return err
		}
	}
	if tagexist {
		for _, tarsum := range oldtarsumlist {
			i := new(models.Image)
			if exists, err := i.Get(tarsum); err != nil {
				return err
			} else if !exists {
				return fmt.Errorf("blobs not existed")
			}

			i.Count = i.Count - 1
			if i.Count == 0 {
				if err := i.Delete(tarsum); err != nil {
					return err
				}
				if err := DeleteLayer(tarsum, "layer", setting.APIVERSION_V2); err != nil {
					return err
				}
			} else {
				if err := i.Save(tarsum); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

//Update repository info in db
func UpdateTaglist(namespace, repository, tag string) error {
	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil || !exists {
		return fmt.Errorf("blobs invalid")
	}

	exists := false
	tagslist := r.GetTagslist()
	for k, v := range tagslist {
		if v == tag {
			exists = true
			kk := k + 1
			tagslist = append(tagslist[:k], tagslist[kk:]...)
			break
		}
	}
	if exists == false {
		return fmt.Errorf("no tags")
	}
	if len(tagslist) == 0 {
		if err := r.Delete(namespace, repository); err != nil {
			return err
		}

		return nil
	}

	r.Tagslist = r.SaveTagslist(tagslist)
	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	return nil
}

//if digest of tag accord with the reference, then delete the tag info
func DeleteTagByRefer(namespace, repository, reference string, tagslist []string) error {
	tagexists := false
	for _, tag := range tagslist {
		t := new(models.Tag)
		if exists, err := t.Get(namespace, repository, tag); err != nil {
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
			if err := UpdateTaglist(namespace, repository, tag); err != nil {
				return err
			}

			if tarsumlist, err := GetTarsumlist([]byte(t.Manifest)); err != nil {
				return err
			} else {
				for _, tarsum := range tarsumlist {
					i := new(models.Image)
					if exists, err := i.Get(tarsum); err != nil {
						return err
					} else if !exists {
						return fmt.Errorf("Image is not exists!")
					}

					i.Count = i.Count - 1
					if err := i.Save(tarsum); err != nil {
						return err
					}
				}
			}

			if err := t.Delete(namespace, repository, tag); err != nil {
				return err
			}
		}
	}
	if tagexists == false {
		return fmt.Errorf("Tag is not exists!")
	}

	return nil
}

//Upload the layer of image to object storage service,support to analyzed docker V1/V2 manifest now
func UploadLayer(tarsumlist []string) error {
	if backend.Drv == nil {
		return nil
	}

	if len(tarsumlist) <= 0 {
		return fmt.Errorf("no blobs")
	}

	var pathlist []string
	var issuccess bool = true
	var err error
	for _, tarsum := range tarsumlist {
		i := new(models.Image)

		var exists bool
		if exists, err = i.Get(tarsum); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("layer is not existent")
		}
		if _, err = os.Stat(i.Path); err != nil && !setting.Cachable {
			continue
		}

		pathlist = append(pathlist, i.Path)
		//TODO: consider to solve saving same layer mutiple times,different from each OSS
		if _, err = backend.Drv.Save(i.Path); err != nil {
			issuccess = false
			break
		}
	}

	//Remove the layer in local fs while upload successfully
	if !setting.Cachable {
		for _, v := range tarsumlist {
			CleanCache(v, setting.APIVERSION_V2)
		}
	}

	//Remove the layer in oss while upload failed
	if !issuccess {
		for _, v := range pathlist {
			backend.Drv.Delete(v)
		}

		for _, v := range tarsumlist {
			i := new(models.Image)
			i.Get(v)
			i.Delete(v)
		}
		return err
	}

	return nil
}

func DownloadLayer(layerpath string) (*os.File, error) {
	var err error
	if _, err := os.Stat(layerpath); err == nil {
		if fs, err := os.Open(layerpath); err == nil {
			return fs, nil
		}
	}

	if backend.Drv == nil {
		return nil, fmt.Errorf("fail to read file: %v", err.Error())
	}

	fs, err := backend.Drv.ReadStream(layerpath, 0)
	if err != nil {
		return nil, fmt.Errorf("fail to download layer: %v", err.Error())
	}

	return fs, nil
}

func DeleteLayer(imageId, layerfile string, apiversion int64) error {
	CleanCache(imageId, apiversion)

	if backend.Drv != nil {
		layerPath := GetLayerPath(imageId, layerfile, apiversion)
		if err := backend.Drv.Delete(layerPath); err != nil {
			return err
		}
	}

	return nil
}

func SaveLayerLocal(srcPath, srcFile, dstPath, dstFile string, reqbody []byte) (int, error) {
	if !utils.IsDirExist(dstPath) {
		os.MkdirAll(dstPath, os.ModePerm)
	}

	if utils.IsFileExist(dstFile) {
		os.Remove(dstFile)
	}

	var data []byte
	if _, err := os.Stat(srcFile); err == nil {
		//docker 1.9.x above version saves layer in PATCH methord
		if err := os.Rename(srcFile, dstFile); err != nil {
			return 0, err
		}
	} else {
		//docker 1.9.x below version saves layer in PUT methord
		data = reqbody
		if err := ioutil.WriteFile(dstFile, data, 0777); err != nil {
			return 0, err
		}
	}

	return len(data), nil
}

//codes as below are ported to support for docker to parse request URL,and it would be update soon
func parseIP(ipStr string) net.IP {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Errorf("Invalid remote IP address: %q", ipStr)
	}
	return ip
}

func RemoteAddr(r *http.Request) string {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		proxies := strings.Split(prior, ",")
		if len(proxies) > 0 {
			remoteAddr := strings.Trim(proxies[0], " ")
			if parseIP(remoteAddr) != nil {
				return remoteAddr
			}
		}
	}

	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		if parseIP(realIP) != nil {
			return realIP
		}
	}

	return r.RemoteAddr
}

const (
	RouteNameBase            = "base"
	RouteNameBlob            = "blob"
	RouteNameManifest        = "manifest"
	RouteNameTags            = "tags"
	RouteNameBlobUpload      = "blob-upload"
	RouteNameBlobUploadChunk = "blob-upload-chunk"
	RouteNameCatalog         = "catalog"
)

type URLBuilder struct {
	root   *url.URL
	router *mux.Router
}

type RouteDescriptor struct {
	Name string
	Path string
}

var RepositoryNameComponentRegexp = regexp.MustCompile(`[a-z0-9]+(?:[._-][a-z0-9]+)*`)
var RepositoryNameRegexp = regexp.MustCompile(`(?:` + RepositoryNameComponentRegexp.String() + `/)*` + RepositoryNameComponentRegexp.String())
var TagNameRegexp = regexp.MustCompile(`[\w][\w.-]{0,127}`)
var DigestRegexp = regexp.MustCompile(`[a-zA-Z0-9-_+.]+:[a-fA-F0-9]+`)

var routeDescriptors = []RouteDescriptor{
	{
		Name: RouteNameBase,
		Path: "/v2/",
	},
	{
		Name: RouteNameBlob,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/{digest:" + DigestRegexp.String() + "}",
	},
	{
		Name: RouteNameManifest,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/manifests/{reference:" + TagNameRegexp.String() + "|" + DigestRegexp.String() + "}",
	},
	{
		Name: RouteNameTags,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/tags/list",
	},
	{
		Name: RouteNameBlobUpload,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/uploads/",
	},
	{
		Name: RouteNameBlobUploadChunk,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/uploads/{uuid:[a-zA-Z0-9-_.=]+}",
	},
}

func NewURLBuilderFromRequest(r *http.Request) *URLBuilder {
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
	/*
		basePath := routeDescriptorsMap[RouteNameBase].Path
		requestPath := r.URL.Path
		index := strings.Index(requestPath, basePath)
		if index > 0 {
			u.Path = requestPath[0 : index+1]
		}
	*/
	return NewURLBuilder(u)
}

func Router() *mux.Router {
	return RouterWithPrefix("")
}

func RouterWithPrefix(prefix string) *mux.Router {
	rootRouter := mux.NewRouter()
	router := rootRouter
	if prefix != "" {
		router = router.PathPrefix(prefix).Subrouter()
	}

	router.StrictSlash(true)

	for _, descriptor := range routeDescriptors {
		router.Path(descriptor.Path).Name(descriptor.Name)
	}

	return rootRouter
}

func NewURLBuilder(root *url.URL) *URLBuilder {
	return &URLBuilder{
		root:   root,
		router: Router(),
	}
}

func (ub *URLBuilder) BuildBlobURL(name string, dgst string) (string, error) {
	route := ub.cloneRoute(RouteNameBlob)

	layerURL, err := route.URL("name", name, "digest", dgst)
	if err != nil {
		return "", err
	}

	return layerURL.String(), nil
}

func (ub *URLBuilder) BuildManifestURL(name, reference string) (string, error) {
	route := ub.cloneRoute(RouteNameManifest)

	manifestURL, err := route.URL("name", name, "reference", reference)
	if err != nil {
		return "", err
	}

	return manifestURL.String(), nil
}

func (ub *URLBuilder) cloneRoute(name string) clonedRoute {
	route := new(mux.Route)
	root := new(url.URL)

	*route = *ub.router.GetRoute(name)
	*root = *ub.root

	return clonedRoute{Route: route, root: root}
}

type clonedRoute struct {
	*mux.Route
	root *url.URL
}

func (cr clonedRoute) URL(pairs ...string) (*url.URL, error) {
	routeURL, err := cr.Route.URL(pairs...)
	if err != nil {
		return nil, err
	}

	if routeURL.Scheme == "" && routeURL.User == nil && routeURL.Host == "" {
		routeURL.Path = routeURL.Path[1:]
	}

	return cr.root.ResolveReference(routeURL), nil
}
