package module

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
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
		fmt.Println("####### SendHttpRequest 0: ", err.Error())
		return &http.Response{}, err
	}

	var client *http.Client
	switch url.Scheme {
	case "":
		fallthrough
	case "https":
		fmt.Println("####### SendHttpRequest 1: ")
		pool := x509.NewCertPool()
		crt, err := ioutil.ReadFile(setting.HttpsCertFile)
		if err != nil {
			fmt.Println("####### SendHttpRequest 2: ", err.Error())
			return &http.Response{}, err
		}
		pool.AppendCertsFromPEM(crt)
		tr := &http.Transport{
			TLSClientConfig:    &tls.Config{RootCAs: pool},
			DisableCompression: true,
		}
		client = &http.Client{Transport: tr}
		fmt.Println("####### SendHttpRequest 3: ")
	case "http":
		//tr := http.DefaultTransport.(*http.Transport)
		client = &http.Client{}
		fmt.Println("####### SendHttpRequest 4: ")
	default:
		return &http.Response{}, fmt.Errorf("wrong url schema: %v", url.Scheme)
	}

	req, err := http.NewRequest(methord, url.String(), body)
	if err != nil {
		fmt.Println("####### SendHttpRequest 2: ", err.Error())
		return &http.Response{}, err
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	//req.Header.Set("Authorization", author)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("####### SendHttpRequest 3: ", err.Error())
		return &http.Response{}, err
	}
	//defer resp.Body.Close()
	return resp, nil
}

//TODO: 考虑并发情况，同步过程中有push或pull操作
func SaveSynContent(namespace, repository, tag string, reqbody []byte) error {
	sc := new(models.Syncont)
	sc.Layers = make(map[string][]byte)
	if err := json.Unmarshal(reqbody, sc); err != nil {
		return err
	}

	//cover repo
	r := new(models.Repository)
	existed, err := r.Get(namespace, repository)
	if err != nil {
		return err
	}
	//Id,Memo,Created,Updated
	r.Namespace = sc.Repository.Namespace
	r.Repository = sc.Repository.Repository
	r.Agent = sc.Repository.Agent
	r.Size = sc.Repository.Size
	r.Version = sc.Repository.Version
	if !existed {
		r.Tagslist = r.SaveTagslist([]string{tag})
		r.JSON = sc.Repository.JSON
		r.Download = 0
	} else {
		r.JSON = "" //TODO
		r.Tagslist = r.SaveTagslist([]string{tag})
		exists := false
		tagslist := r.GetTagslist()
		for _, v := range tagslist {
			if v == tag {
				exists = true
				break
			}
		}
		if !exists {
			tagslist = append(tagslist, tag)
		}
		r.Tagslist = r.SaveTagslist(tagslist)
		//删除原来的metadata
		//TODO
	}
	if err := r.Save(r.Namespace, r.Repository); err != nil {
		return err
	}

	//cover tag
	t := new(models.Tag)
	if _, err := t.Get(namespace, repository, tag); err != nil {
		return err
	}
	//Id,Memo,Created,Updated
	t.Namespace = sc.Tag.Namespace
	t.Repository = sc.Tag.Repository
	t.Tag = sc.Tag.Tag
	t.ImageId = sc.Tag.ImageId
	t.Manifest = sc.Tag.Manifest
	t.Schema = sc.Tag.Schema
	if err := t.Save(t.Namespace, t.Repository, t.Tag); err != nil {
		return err
	}

	//cover image
	var tarsumlist = []string{}
	for _, synimg := range sc.Images {
		i := new(models.Image)
		existed, err := i.Get(synimg.ImageId)
		if err != nil {
			return err
		}
		//Id,Memo,Created,Updated,ManiPath,SignPath,AciPath
		i.ImageId = synimg.ImageId
		i.JSON = synimg.JSON
		i.Ancestry = synimg.Ancestry
		i.Checksum = synimg.Checksum
		i.Payload = synimg.Payload
		i.Checksumed = true
		i.Uploaded = true
		i.Path = GetLayerPath(synimg.ImageId, "layer", setting.APIVERSION_V2)
		i.Size = synimg.Size
		i.Version = synimg.Version
		if !existed {
			i.Count = 0
		}

		if err := i.Save(i.ImageId); err != nil {
			return err
		}

		//删除/覆盖本地原来的镜像
		//TODO :save到对象存储端
		//setting.Cachable
		imgpath := GetImagePath(i.ImageId, setting.APIVERSION_V2)
		if !utils.IsDirExist(imgpath) {
			if err := os.MkdirAll(imgpath, os.ModePerm); err != nil {
				return err
			}
		}
		if err := ioutil.WriteFile(i.Path, sc.Layers[i.ImageId], 0777); err != nil {
			return err
		}
		tarsumlist = append(tarsumlist, i.ImageId)
	}

	if err := UploadLayer(tarsumlist); err != nil {
		return err
	}

	return nil
}

func FillSynContent(namespace, repository, tag string, sc *models.Syncont) error {
	r := new(models.Repository)
	if existed, err := r.Get(namespace, repository); err != nil {
		return err
	} else if !existed {
		return fmt.Errorf("not found repository %s/%s", namespace, repository)
	}
	sc.Repository = *r

	t := new(models.Tag)
	if existed, err := t.Get(namespace, repository, tag); err != nil {
		return err
	} else if !existed {
		return fmt.Errorf("not found tag %s/%s:%s", namespace, repository, tag)
	}
	sc.Tag = *t

	//analyze manifest and get image metadate
	tarsumlist, err := GetTarsumlist([]byte(t.Manifest))
	if err != nil {
		return err
	}
	//get all images metadata
	for _, imageId := range tarsumlist {
		i := new(models.Image)
		if exists, err := i.Get(imageId); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("not found image %s", imageId)
		}
		sc.Images = append(sc.Images, *i)

		//get layer from local or OSS
		if data, err := DownloadLayer(i.Path); err != nil {
			return err
		} else {
			sc.Layers[imageId] = data
		}
	}

	//models.SynConts = append(models.SynConts, *sc)

	return nil
}

func TrigSynch(namespace, repository, tag, dest string) error {
	sc := new(models.Syncont)
	sc.Layers = make(map[string][]byte)
	if err := FillSynContent(namespace, repository, tag, sc); err != nil {
		return err
	}

	//trigger synchronous distribution immediately
	body, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	rawurl := fmt.Sprintf("%s/syn/%s/%s/%s/content", dest, namespace, repository, tag)
	fmt.Println("####### TrigSynch 0: ")
	if resp, err := SendHttpRequest("PUT", rawurl, bytes.NewReader(body)); err != nil {
		fmt.Println("####### TrigSynch 1: ")
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("response code %v", resp.StatusCode)
	}

	fmt.Println("####### TrigSynch 2: ")
	return nil
}

func SaveRegionContent(namespace, repository, tag string, reqbody []byte) error {
	egin := new(models.Endpointgrp)
	if err := json.Unmarshal(reqbody, egin); err != nil {
		return err
	}

	re := new(models.Region)
	if existed, err := re.Get(namespace, repository, tag); err != nil {
		return err
	} else if !existed {
		for k, _ := range egin.Endpoints {
			egin.Endpoints[k].Active = true
		}
		result, _ := json.Marshal(egin)
		re.Namespace, re.Repository, re.Tag, re.Endpointlist =
			namespace, repository, tag, string(result)
	} else {
		egorig := new(models.Endpointgrp)
		if err := json.Unmarshal([]byte(re.Endpointlist), egorig); err != nil {
			return err
		}

		for _, epin := range egin.Endpoints {
			exists := false
			for k, _ := range egorig.Endpoints {
				if epin.URL == egorig.Endpoints[k].URL {
					exists = true
					egorig.Endpoints[k].Active = true
					break
				}
			}

			if !exists {
				epin.Active = true
				egorig.Endpoints = append(egorig.Endpoints, epin)
			}
		}

		result, _ := json.Marshal(egorig)
		re.Endpointlist = string(result)
	}

	if err := re.Save(namespace, repository, tag); err != nil {
		return err
	}
	//TODO: mutex
	models.Regions = append(models.Regions, *re)

	return nil
}

func TrigSynEndpoint(region *models.Region) error {
	epg := new(models.Endpointgrp)
	if err := json.Unmarshal([]byte(region.Endpointlist), epg); err != nil {
		return err
	}

	errCnt := 0
	for k, _ := range epg.Endpoints {
		if epg.Endpoints[k].Active == false {
			continue
		}
		//TODO: goroutine
		if err := TrigSynch(region.Namespace, region.Repository, region.Tag, epg.Endpoints[k].URL); err != nil {
			errCnt++
			fmt.Printf("[REGISTRY API] Failed to synchronize %s: %s", epg.Endpoints[k].URL, err.Error())
			continue
		} else {
			epg.Endpoints[k].Active = false
		}
	}

	if len(epg.Endpoints) == errCnt {
		return fmt.Errorf("all endpoints synchronized failed")
	}

	result, _ := json.Marshal(epg)
	region.Endpointlist = string(result)
	if err := region.Save(region.Namespace, region.Repository, region.Tag); err != nil {
		return err
	}

	if errCnt != 0 { //TODO
		return fmt.Errorf("some synchron failed")
	}

	return nil
}
