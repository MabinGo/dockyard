package synch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	RTName  string = "RegionTable"
	SUCCESS string = "Success"
	FAILURE string = "Failure"
	ORIGION string = "Origin"
)

func TrigSynDRC(namespace, repository, tag, auth string) error {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("region table not found")
	}

	if rt.DRClist != "" {
		eplist := new(Endpointlist)
		if err := json.Unmarshal([]byte(rt.DRClist), eplist); err != nil {
			return err
		}

		for k, v := range eplist.Endpoints {
			if v.Active == false {
				continue
			}

			if err := trig(namespace, repository, tag, auth, v.URL); err != nil {
				synlog.Error("Failed to synchronize %s/%s:%s to DR %s, error: %v", namespace, repository, tag, v.URL, err)
				eplist.Endpoints[k].Status = FAILURE
			} else {
				synlog.Trace("Successed to synchronize %s/%s:%s to DR %s", namespace, repository, tag, v.URL)
				eplist.Endpoints[k].Status = SUCCESS
			}
		}

		result, _ := json.Marshal(eplist)
		rt.DRClist = string(result)
		if err := rt.Save(RTName); err != nil {
			return err
		}
	}

	return nil
}

func TrigSynEndpoint(region *Region, auth string) error {
	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(region.Endpointlist), eplist); err != nil {
		return err
	}

	activecnt := 0
	errs := []string{}
	for k := range eplist.Endpoints {
		if eplist.Endpoints[k].Active == false {
			activecnt++
			continue
		}
		//TODO: opt to use goroutine
		if err := trig(region.Namespace, region.Repository, region.Tag, auth, eplist.Endpoints[k].URL); err != nil {
			//synlog.Error("Failed to synchronize %s/%s:%s to %s, error: %v",
			//	region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL, err)
			errs = append(errs, fmt.Sprintf("synchronize to %s error: %s", eplist.Endpoints[k].URL, err.Error()))
			eplist.Endpoints[k].Status = FAILURE
			continue
		} else {
			//synlog.Trace("Successed to synchronize %s/%s:%s to %s",
			//	region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL)
			eplist.Endpoints[k].Status = SUCCESS
		}
	}

	if activecnt == len(eplist.Endpoints) {
		synlog.Trace("no active region")
	}

	result, _ := json.Marshal(eplist)
	region.Endpointlist = string(result)
	if err := region.Save(region.Namespace, region.Repository, region.Tag); err != nil {
		return err
	}
	/*
		if len(eplist.Endpoints) == len(errs) {
			return fmt.Errorf("%v", errs)
		}
	*/
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

func trig(namespace, repository, tag, auth, dest string) error {
	sc := new(Syncont)
	sc.Layers = make(map[string][]byte)
	if err := fillSynContent(namespace, repository, tag, sc); err != nil {
		return err
	}

	//trigger synchronous distribution immediately
	body, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/syn/%s/%s/%s/content", dest, namespace, repository, tag)
	if resp, err := module.SendHttpRequest("PUT", url, bytes.NewReader(body), auth); err != nil {
		//synlog.Error("Failed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
		return err
	} else if resp.StatusCode != http.StatusOK { //http.StatusInternalServerError
		return fmt.Errorf("response code %v", resp.StatusCode)
		//synlog.Error("Failed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
	}

	return nil
}

//TODO: must consider parallel, push/pull during synchron
func SaveSynContent(namespace, repository, tag string, reqbody []byte) error {
	sc := new(Syncont)
	sc.Layers = make(map[string][]byte)
	if err := json.Unmarshal(reqbody, sc); err != nil {
		return err
	}

	//cover repo
	r := new(models.Repository)
	exists, err := r.Get(namespace, repository)
	if err != nil {
		return err
	}
	//Id,Memo,Created,Updated
	r.Namespace = sc.Repository.Namespace
	r.Repository = sc.Repository.Repository
	r.Agent = sc.Repository.Agent
	r.Size = sc.Repository.Size
	r.Version = sc.Repository.Version
	r.JSON = sc.Repository.JSON
	if !exists {
		r.Tagslist = r.SaveTagslist([]string{tag})
		r.Download = 0
	} else {
		//r.Tagslist = r.SaveTagslist([]string{tag})
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
		exists, err := i.Get(synimg.ImageId)
		if err != nil {
			return err
		}
		//Id,Memo,Created,Updated,ManiPath,SignPath,AciPath
		i.ImageId = synimg.ImageId
		i.JSON = synimg.JSON
		i.Ancestry = synimg.Ancestry
		i.Checksum = synimg.Checksum
		i.Payload = synimg.Payload
		i.Checksumed = synimg.Checksumed
		i.Uploaded = synimg.Uploaded
		i.Path = module.GetLayerPath(synimg.ImageId, "layer", setting.APIVERSION_V2)
		i.Size = synimg.Size
		i.Version = synimg.Version
		if !exists {
			i.Count = 0
		}

		if err := i.Save(i.ImageId); err != nil {
			return err
		}

		//TODO: consider to delete or cover the origin images
		imgpath := module.GetImagePath(i.ImageId, setting.APIVERSION_V2)
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

	//upload to oss
	if err := module.UploadLayer(tarsumlist); err != nil {
		return err
	}

	return nil
}

func fillSynContent(namespace, repository, tag string, sc *Syncont) error {
	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found repository %s/%s:%s", namespace, repository, tag)
	}
	sc.Repository = *r

	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found tag %s/%s:%s", namespace, repository, tag)
	}
	sc.Tag = *t

	//analyze manifest and get image metadate
	tarsumlist, err := module.GetTarsumlist([]byte(t.Manifest))
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
		if data, err := module.DownloadLayer(i.Path); err != nil {
			return err
		} else {
			sc.Layers[imageId] = data
		}
	}

	return nil
}

func SaveRegionContent(namespace, repository, tag string, reqbody []byte) error {
	eplist := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplist); err != nil {
		return err
	}

	regionIn := new(Region)
	if exists, err := regionIn.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exists {
		for k := range eplist.Endpoints {
			eplist.Endpoints[k].Status = ORIGION
		}
		result, _ := json.Marshal(eplist)
		regionIn.Namespace, regionIn.Repository, regionIn.Tag, regionIn.Endpointlist =
			namespace, repository, tag, string(result)
	} else {
		eporig := new(Endpointlist)
		if err := json.Unmarshal([]byte(regionIn.Endpointlist), eporig); err != nil {
			return err
		}

		for _, epin := range eplist.Endpoints {
			exists := false
			for k, v := range eporig.Endpoints {
				if epin.URL == v.URL {
					exists = true
					eporig.Endpoints[k] = epin
					eporig.Endpoints[k].Status = ORIGION
					break
				}
			}

			if !exists {
				epin.Status = ORIGION
				eporig.Endpoints = append(eporig.Endpoints, epin)
			}
		}

		result, _ := json.Marshal(eporig)
		regionIn.Endpointlist = string(result)
	}

	if err := regionIn.Save(namespace, repository, tag); err != nil {
		return err
	}

	//TODO: mutex
	if setting.SynMode != "" {
		if err := UpdateRegionList(regionIn); err != nil {
			return err
		}
	}

	return nil
}

func SaveContent(regiontyp string, reqbody []byte) error {
	eplistIn := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplistIn); err != nil {
		return err
	}

	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("region table not found")
	}

	var list string
	eplistOri := new(Endpointlist)
	switch regiontyp {
	case MASTER:
		if rt.Masterlist != "" {
			if err := json.Unmarshal([]byte(rt.Masterlist), eplistOri); err != nil {
				return err
			}
		}
		list = rt.Masterlist
	case DRC:
		if rt.DRClist != "" {
			if err := json.Unmarshal([]byte(rt.DRClist), eplistOri); err != nil {
				return err
			}
		}
		list = rt.DRClist
	case COMMON:
	default:
		return fmt.Errorf("not support region type")
	}

	if list != "" {
		for _, epin := range eplistIn.Endpoints {
			exists := false
			for k, v := range eplistOri.Endpoints {
				if epin.URL == v.URL {
					exists = true
					eplistOri.Endpoints[k] = epin
					eplistOri.Endpoints[k].Status = ORIGION
					break
				}
			}

			if !exists {
				epin.Status = ORIGION
				eplistOri.Endpoints = append(eplistOri.Endpoints, epin)
			}
		}
	} else {
		eplistOri = eplistIn
		for k := range eplistOri.Endpoints {
			eplistOri.Endpoints[k].Status = ORIGION
		}
	}

	result, _ := json.Marshal(eplistOri)
	if regiontyp == MASTER {
		rt.Masterlist = string(result)
	} else if regiontyp == DRC {
		rt.DRClist = string(result)
	} else {
	}

	if err := rt.Save(RTName); err != nil {
		return err
	}

	return nil
}

func UpdateRegionList(regionIn *Region) error {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("region table invalid")
	}

	rl := new(Regionlist)
	if rt.Regionlist != "" {
		if err := json.Unmarshal([]byte(rt.Regionlist), rl); err != nil {
			return err
		}

		exists := false
		index := 0
		for k, v := range rl.Regions {
			if v.Id == regionIn.Id {
				exists = true
				index = k
				break
			}
		}

		if !exists {
			rl.Regions = append(rl.Regions, *regionIn)
		} else {
			rl.Regions[index] = *regionIn
		}
	} else {
		rl.Regions = append(rl.Regions, *regionIn)
	}
	result, _ := json.Marshal(rl)
	rt.Regionlist = string(result)

	if err := rt.Save(RTName); err != nil {
		return err
	}

	return nil
}

func GetSynList(regiontyp string) (string, error) {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("region table invalid")
	}

	var list string
	if regiontyp == MASTER {
		list = rt.Masterlist
	} else if regiontyp == DRC {
		list = rt.DRClist
	}

	return list, nil
}

func GetSynRegionEndpoint(namespace, repository, tag string) (string, error) {
	r := new(Region)
	if exists, err := r.Get(namespace, repository, tag); err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	return r.Endpointlist, nil
}

func DelSynRegion(namespace, repository, tag string, reqbody []byte) (bool, error) {
	eplistIn := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplistIn); err != nil {
		return false, err
	}

	r := new(Region)
	if exists, err := r.Get(namespace, repository, tag); err != nil {
		return false, err
	} else if !exists {
		return false, fmt.Errorf("not found region")
	}

	if len(r.Endpointlist) <= 0 {
		return false, fmt.Errorf("endpoint list is null")
	}

	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(r.Endpointlist), eplist); err != nil {
		return false, err
	}

	orilen := len(eplist.Endpoints)
	for _, epIn := range eplistIn.Endpoints {
		for k, v := range eplist.Endpoints {
			if epIn.URL == v.URL {
				eplist.Endpoints[k].URL = ""
				continue
			}
		}
	}

	eplistNew := new(Endpointlist)
	for _, ep := range eplist.Endpoints {
		if ep.URL == "" {
			continue
		}
		eplistNew.Endpoints = append(eplistNew.Endpoints, ep)
	}

	newlen := len(eplistNew.Endpoints)
	if newlen == 0 {
		return true, r.Delete(namespace, repository, tag)
	}

	if newlen == orilen {
		return false, nil
	}

	result, _ := json.Marshal(eplistNew)
	r.Endpointlist = string(result)

	if err := r.Save(namespace, repository, tag); err != nil {
		return false, err
	}

	//TODO: mutex
	if setting.SynMode != "" {
		if err := UpdateRegionList(r); err != nil {
			return false, err
		}
	}

	return true, nil
}

func DelSynEndpoint(regiontyp string, reqbody []byte) (bool, error) {
	eplistIn := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplistIn); err != nil {
		return false, err
	}

	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return false, err
	} else if !exists {
		return false, fmt.Errorf("not found region table")
	}

	eplist := new(Endpointlist)
	switch regiontyp {
	case MASTER:
		if len(rt.Masterlist) <= 0 {
			return false, fmt.Errorf("Master list is null")
		}

		if err := json.Unmarshal([]byte(rt.Masterlist), eplist); err != nil {
			return false, err
		}
	case DRC:
		if len(rt.DRClist) <= 0 {
			return false, fmt.Errorf("DRC list is null")
		}

		if err := json.Unmarshal([]byte(rt.DRClist), eplist); err != nil {
			return false, err
		}
	case COMMON:
	default:
		return false, fmt.Errorf("not support region type")
	}

	orilen := len(eplist.Endpoints)
	for _, epIn := range eplistIn.Endpoints {
		for k, v := range eplist.Endpoints {
			if epIn.URL == v.URL {
				eplist.Endpoints[k].URL = ""
				continue
			}
		}
	}

	eplistNew := new(Endpointlist)
	for _, ep := range eplist.Endpoints {
		if ep.URL == "" {
			continue
		}
		eplistNew.Endpoints = append(eplistNew.Endpoints, ep)
	}

	newlen := len(eplistNew.Endpoints)
	if newlen == 0 {
		if regiontyp == MASTER {
			rt.Masterlist = ""
		} else if regiontyp == DRC {
			rt.DRClist = ""
		}
		return true, rt.Save(RTName)
	}

	if newlen == orilen {
		return false, nil
	}

	result, _ := json.Marshal(eplistNew)
	if regiontyp == MASTER {
		rt.Masterlist = string(result)
	} else if regiontyp == DRC {
		rt.DRClist = string(result)
	}

	if err := rt.Save(RTName); err != nil {
		return false, err
	}

	return true, nil
}

func IsMasterExisted() bool {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil || !exists {
		return false
	}

	if rt.Masterlist != "" {
		return true
	} else {
		return false
	}
}

func GetSynFromMaster(namespace, repository, tag, auth string) error {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found region table")
	}

	if rt.Masterlist == "" {
		return fmt.Errorf("no remote endpoint")
	}

	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(rt.Masterlist), eplist); err != nil {
		return err
	}

	for _, v := range eplist.Endpoints {
		url := fmt.Sprintf("%s/syn/%s/%s/%s/content", v.URL, namespace, repository, tag)
		resp, err := module.SendHttpRequest("GET", url, nil, auth)
		if err != nil {
			continue
		} else if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		if err := SaveSynContent(namespace, repository, tag, body); err != nil {
			continue
		}

		return nil
	}

	return fmt.Errorf("bad remote endpoint")
}

func GetTaglistFromMaster(namespace, repository, auth string) ([]string, error) {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return []string{}, err
	} else if !exists {
		return []string{}, fmt.Errorf("not found region table")
	}

	if rt.Masterlist == "" {
		return []string{}, fmt.Errorf("no remote endpoint")
	}

	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(rt.Masterlist), eplist); err != nil {
		return []string{}, err
	}

	for _, v := range eplist.Endpoints {
		url := fmt.Sprintf("%s/syn/%s/%s/tags/list", v.URL, namespace, repository)
		resp, err := module.SendHttpRequest("GET", url, nil, auth)
		if err != nil {
			continue
		} else if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var tags = []string{}
		if err := json.Unmarshal(body, &tags); err != nil {
			return []string{}, err
		}

		return tags, nil
	}

	return []string{}, fmt.Errorf("bad remote endpoint")
}
