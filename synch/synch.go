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

var RTName string = "RegionTable"

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

		for _, v := range eplist.Endpoints {
			if v.Active == false {
				continue
			}

			if err := trig(namespace, repository, tag, auth, v.URL); err != nil {
				synlog.Error("\nFailed to synchronize %s/%s:%s to DR %s, error: %v", namespace, repository, tag, v.URL, err)
			} else {
				synlog.Trace("\nSuccessed to synchronize %s/%s:%s to DR %s", namespace, repository, tag, v.URL)
			}
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
	for k, _ := range eplist.Endpoints {
		if eplist.Endpoints[k].Active == false {
			activecnt++
			continue
		}
		//TODO: opt to use goroutine
		if err := trig(region.Namespace, region.Repository, region.Tag, auth, eplist.Endpoints[k].URL); err != nil {
			synlog.Error("\nFailed to synchronize %s/%s:%s to %s, error: %v",
				region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL, err)
			errs = append(errs, fmt.Sprintf("\nsynchronize to %s error: %s", eplist.Endpoints[k].URL, err.Error()))
			continue
		} else {
			synlog.Trace("\nSuccessed to synchronize %s/%s:%s to %s",
				region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL)
		}
	}

	if activecnt == len(eplist.Endpoints) {
		synlog.Trace("\nno active region")
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
	if err := FillSynContent(namespace, repository, tag, sc); err != nil {
		return err
	}

	//trigger synchronous distribution immediately
	body, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/syn/%s/%s/%s/content", dest, namespace, repository, tag)

	var times = 10
	var ret error
	for i := times; i > 0; i-- {
		rawurl := fmt.Sprintf("%s?times=%v&count=%v", url, times, i)
		if resp, err := module.SendHttpRequest("PUT", rawurl, bytes.NewReader(body), auth); err != nil {
			//synlog.Error("\nFailed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
			ret = err
			break
		} else if resp.StatusCode != http.StatusOK { //http.StatusInternalServerError
			ret = fmt.Errorf("response code %v", resp.StatusCode)
			//synlog.Error("\nFailed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
			continue
		} else {
			//TODO: must announce success to user
			//synlog.Trace("\nSuccess synchronize %s/%s:%s to %s", namespace, repository, tag, dest)
			ret = nil
			break
		}
	}

	return ret
}

//TODO: must consider parallel, push/pull during synchron
func SaveSynContent(namespace, repository, tag, count string, reqbody []byte) error {
	sc := new(Syncont)
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
	r.JSON = sc.Repository.JSON
	if !existed {
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
		i.Checksumed = synimg.Checksumed
		i.Uploaded = synimg.Uploaded
		i.Path = module.GetLayerPath(synimg.ImageId, "layer", setting.APIVERSION_V2)
		i.Size = synimg.Size
		i.Version = synimg.Version
		if !existed {
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

func FillSynContent(namespace, repository, tag string, sc *Syncont) error {
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
	if existed, err := regionIn.Get(namespace, repository, tag); err != nil {
		return err
	} else if !existed {
		//for k, _ := range eplist.Endpoints {
		//	eplist.Endpoints[k].Active = true
		//}
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
					//eporig.Endpoints[k].Active = true
					break
				}
			}

			if !exists {
				//epin.Active = true
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

func SaveDRCContent(reqbody []byte) error {
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

	eplistOri := new(Endpointlist)
	if rt.DRClist != "" {
		if err := json.Unmarshal([]byte(rt.DRClist), eplistOri); err != nil {
			return err
		}

		for _, epin := range eplistIn.Endpoints {
			exists := false
			for k, v := range eplistOri.Endpoints {
				if epin.URL == v.URL {
					exists = true
					eplistOri.Endpoints[k] = epin
					//eplistOri.Endpoints[k].Active = true
					break
				}
			}

			if !exists {
				//epin.Active = true
				eplistOri.Endpoints = append(eplistOri.Endpoints, epin)
			}
		}
	} else {
		eplistOri = eplistIn
		//for k, _ := range eplistOri.Endpoints {
		//	eplistOri.Endpoints[k].Active = true
		//}
	}

	result, _ := json.Marshal(eplistOri)
	rt.DRClist = string(result)
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

func GetSynDRCList() (string, error) {
	rt := new(RegionTable)
	if exists, err := rt.Get(RTName); err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("region table invalid")
	}

	return rt.DRClist, nil
}

func GetSynRegionEndpoint(namespace, repository, tag string) (string, error) {
	r := new(Region)
	if exists, err := r.Get(namespace, repository, tag); err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("not found")
	}

	return r.Endpointlist, nil
}

func DelSynRegion(namespace, repository, tag string, reqbody []byte) error {
	eplistIn := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplistIn); err != nil {
		return err
	}

	r := new(Region)
	if existed, err := r.Get(namespace, repository, tag); err != nil {
		return err
	} else if !existed {
		return fmt.Errorf("not found region")
	}

	if len(r.Endpointlist) <= 0 {
		return fmt.Errorf("endpoint list invalid")
	}

	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(r.Endpointlist), eplist); err != nil {
		return err
	}

	eplistNew := new(Endpointlist)

	for _, ep := range eplist.Endpoints {
		//exists := false
		for k, v := range eplistIn.Endpoints {
			if ep.URL != v.URL {
				exists = true
				eplist.Endpoints[k] = epin
				break
			}

			if epin.URL == v.URL {
				exists = true
				eplist.Endpoints[k] = epin
				break
			}
		}

		if !exists {
			eplist.Endpoints = append(eplist.Endpoints, epin)
		}
	}

	result, _ := json.Marshal(eplist)
	r.Endpointlist = string(result)

	if err := r.Save(namespace, repository, tag); err != nil {
		return err
	}

	//TODO: mutex
	if setting.SynMode != "" {
		if err := UpdateRegionList(r); err != nil {
			return err
		}
	}

	return nil
}
