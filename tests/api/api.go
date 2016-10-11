/*
Copyright 2016 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Mode 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Return struct {
	Token     string    `json:"token" description:"token return to user"`
	ExpiresIn int       `json:"expires_in" description:"describes token  will expires in how many seconds later"`
	IssuedAt  time.Time `json:"issued_at" description:"token issued time"`
}

func DockyardHealth() bool {
	if !dockyardHealth {
		resp, err := SendHttpRequest("GET", fmt.Sprintf("%s/%s", DockyardURI, "v2/"), nil, map[string]string{})
		if err == nil && resp.StatusCode == 200 {
			dockyardHealth = true
		} else {
			if err == nil && resp.StatusCode == 401 {
				openAuth = true
				dockyardHealth = true
			} else {
				fmt.Printf("Connect to dockyard faild, error: %v", err)
			}
		}
	}
	return dockyardHealth
}

func CheckDockyardAuth() error {
	if openAuth != AuthEnable {
		return fmt.Errorf("Dockyard auth enable is %s, but test auth enable is %s", openAuth, AuthEnable)
	}
	return nil
}

func SendHttpRequestRepeat(method, rawurl string, body io.Reader, header map[string]string) (*http.Response, error) {
	var err error
	var resp *http.Response
	for i := 0; i < HTTPReconnectionCount; i++ {
		if resp, err = SendHttpRequest(method, rawurl, body, header); err == nil {
			return resp, nil
		}
	}
	return &http.Response{}, err
}

func SendHttpRequest(method, rawurl string, body io.Reader, header map[string]string) (*http.Response, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return &http.Response{}, err
	}

	var client *http.Client
	switch url.Scheme {
	case "":
		fallthrough
	case "https":
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	case "http":
		client = &http.Client{}
	default:
		return &http.Response{}, fmt.Errorf("bad url schema: %v", url.Scheme)
	}

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return &http.Response{}, err
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	for k, v := range header {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}

func GetAuthorize(username, password string) (string, error) {
	url := ListenMode + "://" + Domains + "/uam/auth"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, password)
	if req.ParseForm(); err != nil {
		return "", err
	}

	var client *http.Client
	if ListenMode == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		tr := &http.Transport{
			DisableKeepAlives: true,
		}
		client = &http.Client{Transport: tr}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var token Return
	if err = json.Unmarshal(body, &token); err != nil {
		return "", err
	}

	return token.Token, nil
}
