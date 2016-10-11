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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/containerops/dockyard/utils"
)

type AppV1Repo struct {
	URI        string
	Namespace  string
	Repository string

	host string
}

type AppV1App struct {
	OS   string
	Arch string
	App  string
	Tag  string
}

func NewAppV1App(os, arch, app, tag string) (AppV1App, error) {
	if os == "" || arch == "" || app == "" {
		return AppV1App{}, errors.New("OS, Arch, App should not be empty")
	}

	var v1App AppV1App
	v1App.OS = os
	v1App.Arch = arch
	v1App.App = app
	v1App.Tag = tag

	return v1App, nil
}

func NewAppV1Repo(uri, n, r string) (AppV1Repo, error) {
	if uri == "" {
		return AppV1Repo{}, errors.New("URI should not be empty")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return AppV1Repo{}, err
	}

	var o AppV1Repo
	o.URI = uri
	o.Namespace = n
	o.Repository = r
	o.host = u.Host
	return o, nil
}

func (o *AppV1Repo) pullFileByName(rawurl, token string, fileName string) (string, int, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	header := map[string]string{
		"Host":          o.host,
		"Authorization": token,
	}

	resp, err := SendHttpRequestRepeat("GET", rawurl, nil, header)
	if err != nil {
		fmt.Println(err, "pull SendHttpRequestRepeat")
		return "", 0, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		fmt.Println(err, "pull copy")
		return "", 0, err
	}
	file.Seek(0, 0)
	sha512Sum, err, _ := utils.SHA512FromStream(file)
	if err != nil {
		fmt.Println(err, "pull SHA512FromStream")
		return "", 0, err
	}
	return sha512Sum, resp.StatusCode, nil
}

func (o *AppV1Repo) putFileByName(rawurl, token, uuid string, fileName string) (string, int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	sha512Sum, err, _ := utils.SHA512FromStream(file)
	if err != nil {
		fmt.Println(err, "put SHA512FromStream")
		return "", 0, err
	}
	digest := fmt.Sprintf("%s:%s", "sha512", sha512Sum)
	file.Seek(0, 0)

	header := map[string]string{
		"Host":            o.host,
		"Authorization":   token,
		"App-Upload-UUID": uuid,
		"Digest":          digest,
	}

	resp, err := SendHttpRequestRepeat("PUT", rawurl, file, header)
	if err != nil {
		fmt.Println(err, "put SendHttpRequestRepeat")
		return "", 0, err
	}
	defer resp.Body.Close()
	return sha512Sum, resp.StatusCode, nil
}

func (o *AppV1Repo) SearchGlobal(params []string, token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/search", o.URI)
	if params != nil && len(params) > 0 {
		rawurl = fmt.Sprintf("%s?key=%s", rawurl, strings.Join(params, "+"))
	}

	return o.Get(rawurl, token)
}

func (o *AppV1Repo) SearchScoped(params []string, token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/search", o.URI, o.Namespace, o.Repository)
	if params != nil && len(params) > 0 {
		rawurl = fmt.Sprintf("%s?key=%s", rawurl, strings.Join(params, "+"))
	}
	return o.Get(rawurl, token)
}

func (o *AppV1Repo) GetList(token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/list", o.URI, o.Namespace, o.Repository)

	return o.Get(rawurl, token)
}

func (o *AppV1Repo) GetMeta(token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/meta", o.URI, o.Namespace, o.Repository)

	return o.Get(rawurl, token)
}

func (o *AppV1Repo) GetMetaSign(token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/metasign", o.URI, o.Namespace, o.Repository)

	return o.Get(rawurl, token)
}

func (o *AppV1Repo) GetPublicKey(token string) ([]byte, int, error) {
	rawurl := fmt.Sprintf("%s/key/v1", o.URI)

	return o.Get(rawurl, token)
}

func (o *AppV1Repo) PullFile(v1App AppV1App, token string, fileName string) (string, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}
	return o.pullFileByName(rawurl, token, fileName)
}

func (o *AppV1Repo) PullManifest(v1App AppV1App, token string, fileName string) (string, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s/manifests", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}

	return o.pullFileByName(rawurl, token, fileName)
}

func (o *AppV1Repo) PutFile(v1App AppV1App, token, uuid string, filePath string) (string, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}
	return o.putFileByName(rawurl, token, uuid, filePath)
}

func (o *AppV1Repo) PutManifest(v1App AppV1App, token, uuid string, filePath string) (string, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s/manifests", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}

	return o.putFileByName(rawurl, token, uuid, filePath)
}

func (o *AppV1Repo) Get(rawurl, token string) ([]byte, int, error) {
	header := map[string]string{
		"Host":          o.host,
		"Authorization": token,
	}

	resp, err := SendHttpRequestRepeat("GET", rawurl, nil, header)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}

func (o *AppV1Repo) Post(token string) (string, int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s", o.URI, o.Namespace, o.Repository)
	header := map[string]string{
		"Host":          o.host,
		"Authorization": token,
	}
	resp, err := SendHttpRequestRepeat("POST", rawurl, nil, header)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	return resp.Header.Get("App-Upload-Uuid"), resp.StatusCode, nil
}

func (o *AppV1Repo) Patch(v1App AppV1App, token, uuid, status string) (int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s/%s", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App, status)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}

	header := map[string]string{
		"Host":            o.host,
		"Authorization":   token,
		"App-Upload-UUID": uuid,
	}
	resp, err := SendHttpRequestRepeat("PATCH", rawurl, nil, header)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (o *AppV1Repo) Delete(v1App AppV1App, token string) (int, error) {
	rawurl := fmt.Sprintf("%s/app/v1/%s/%s/%s/%s/%s", o.URI, o.Namespace, o.Repository, v1App.OS, v1App.Arch, v1App.App)
	if v1App.Tag != "" {
		rawurl = fmt.Sprintf("%s/%s", rawurl, v1App.Tag)
	}

	header := map[string]string{
		"Host":          o.host,
		"Authorization": token,
	}
	resp, err := SendHttpRequestRepeat("DELETE", rawurl, nil, header)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
