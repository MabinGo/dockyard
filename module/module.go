package module

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/containerops/dockyard/utils/setting"
)

var Apis = []string{"images", "tarsum", "acis"}

func CleanCache(imageId string, apiversion int64) {
	imagepath := GetImagePath(imageId, apiversion)
	os.RemoveAll(imagepath)
}

func GetPubkeysPath(namespace, repository string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/pubkeys/%v/%v", setting.ImagePath, Apis[apiversion], namespace, repository)
}

func GetImagePath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v", setting.ImagePath, Apis[apiversion], imageId)
}

func GetManifestPath(imageId string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/manifest", setting.ImagePath, Apis[apiversion], imageId)
}

func GetSignaturePath(imageId, signfile string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/%v", setting.ImagePath, Apis[apiversion], imageId, signfile)
}

func GetLayerPath(imageId, layerfile string, apiversion int64) string {
	return fmt.Sprintf("%v/%v/%v/%v", setting.ImagePath, Apis[apiversion], imageId, layerfile)
}

func SendHttpRequest(methord, rawurl string, body io.Reader) (*http.Response, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return &http.Response{}, err
	}

	var client *http.Client
	switch url.Scheme {
	case "":
		fallthrough
	case "https":
		pool := x509.NewCertPool()
		crt, err := ioutil.ReadFile(setting.HttpsCertFile)
		if err != nil {
			return &http.Response{}, err
		}
		pool.AppendCertsFromPEM(crt)
		tr := &http.Transport{
			TLSClientConfig:    &tls.Config{RootCAs: pool},
			DisableCompression: true,
		}
		client = &http.Client{Transport: tr}
	case "http":
		//tr := http.DefaultTransport.(*http.Transport)
		client = &http.Client{}
	default:
		return &http.Response{}, fmt.Errorf("wrong url schema: %v", url.Scheme)
	}

	req, err := http.NewRequest(methord, url.String(), body)
	if err != nil {
		return &http.Response{}, err
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	//req.Header.Set("Authorization", author)

	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}
	//defer resp.Body.Close()
	return resp, nil
}
