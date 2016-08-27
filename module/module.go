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
	"strings"

	"github.com/containerops/dockyard/utils/setting"
)

var Apis = []string{"images", "tarsum", "acis"}

func CleanCache(imageId string, apiversion int64) {
	imagepath := GetImagePath(imageId, apiversion)
	os.RemoveAll(imagepath)
}

func GetTmpFile(name string) string {
	return fmt.Sprintf("%v/tmp/%v", setting.ImagePath, name)
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

func SendHttpRequest(methord, rawurl string, body io.Reader, auth string) (*http.Response, error) {
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
			TLSClientConfig: &tls.Config{
				RootCAs:            pool,
				InsecureSkipVerify: true,
			},
		}
		client = &http.Client{Transport: tr}
	case "http":
		client = &http.Client{}
	default:
		return &http.Response{}, fmt.Errorf("bad url schema: %v", url.Scheme)
	}

	req, err := http.NewRequest(methord, url.String(), body)
	if err != nil {
		return &http.Response{}, err
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	req.Header.Set("Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

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
		// According to the Apache mod_proxy docs, X-Forwarded-Host can be a
		// comma-separated list of hosts, to which each proxy appends the
		// requested host. We want to grab the first from this comma-separated
		// list.
		hosts := strings.SplitN(forwardedHost, ",", 2)
		host = strings.TrimSpace(hosts[0])
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
	}

	return u
}
