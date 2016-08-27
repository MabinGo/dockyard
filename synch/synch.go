package synch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	RTName   = "RegionTable"
	SUCCESS  = "Success"
	FAILURE  = "Failure"
	SYNCHING = "Synching"
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

		for _, v := range eplist.Endpoints {
			if v.Active == false {
				continue
			}

			if err := trig(namespace, repository, tag, auth, v.URL); err != nil {
				synlog.Error("Failed to synchronize %s/%s:%s to DR %s, error: %v", namespace, repository, tag, v.URL, err)
			} else {
				synlog.Trace("Successed to synchronize %s/%s:%s to DR %s", namespace, repository, tag, v.URL)
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

func trigRegionEndpoint(region *Region, auth string) error {
	eplist := new(Endpointlist)
	if err := json.Unmarshal([]byte(region.Endpointlist), eplist); err != nil {
		return err
	}

	//activecnt := 0
	for k := range eplist.Endpoints {
		go func(k int, eplist *Endpointlist, namespace, repository, tag string) {
			/*
				if eplist.Endpoints[k].Active == false {
					activecnt++
					return
				}
			*/
			endpointstr := Endpointstr{
				Area: eplist.Endpoints[k].Area,
				Name: eplist.Endpoints[k].Name,
				URL:  eplist.Endpoints[k].URL,
			}

			synstate := Synstate{
				Status: SYNCHING,
				Time:   time.Now(),
			}
			if err := saveEndpointstatus(namespace, repository, tag, endpointstr, synstate); err != nil {
				return
			}

			//TODO: opt to use goroutine
			if err := trig(namespace, repository, tag, auth, eplist.Endpoints[k].URL); err != nil {
				synlog.Error("Failed to synchronize %s/%s:%s to %s, error: %v",
					region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL, err)

				synstate := Synstate{
					Status:   FAILURE,
					Response: err.Error(),
					Time:     time.Now(),
				}
				if err := saveEndpointstatus(namespace, repository, tag, endpointstr, synstate); err != nil {
					return
				}
			} else {
				synlog.Trace("Successed to synchronize %s/%s:%s to %s",
					region.Namespace, region.Repository, region.Tag, eplist.Endpoints[k].URL)

				synstate := Synstate{
					Status: SUCCESS,
					Time:   time.Now(),
				}
				if err := saveEndpointstatus(namespace, repository, tag, endpointstr, synstate); err != nil {
					return
				}
			}
		}(k, eplist, region.Namespace, region.Repository, region.Tag)
	}

	//if activecnt == len(eplist.Endpoints) {
	//	synlog.Trace("no active region")
	//}

	return nil
}

func saveEndpointstatus(namespace, repository, tag string, epstr Endpointstr, synstate Synstate) error {
	synendpoint := new(SynEndpoint)
	endpointstrlist := new(Endpointstrlist)
	synstalist := new(Synstatelist)

	if exist, err := synendpoint.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exist {
		synstalist.Synstates = append(synstalist.Synstates, synstate)
		result, _ := json.Marshal(synstalist)
		epstr.Synstatelist = string(result)
		endpointstrlist.Endpointstrs = append(endpointstrlist.Endpointstrs, epstr)
		result, _ = json.Marshal(endpointstrlist)
		synendpoint.Endpointstrlist = string(result)
		if err := synendpoint.Save(namespace, repository, tag); err != nil {
			return err
		}
	} else {
		if err := json.Unmarshal([]byte(synendpoint.Endpointstrlist), endpointstrlist); err != nil {
			return err
		}
		exist := false
		for k, v := range endpointstrlist.Endpointstrs {
			if v.Area == epstr.Area && v.Name == epstr.Name && v.URL == epstr.URL {
				if err := json.Unmarshal([]byte(endpointstrlist.Endpointstrs[k].Synstatelist), synstalist); err != nil {
					return err
				}
				synstalist.Synstates = append(synstalist.Synstates, synstate)
				result, _ := json.Marshal(synstalist)
				endpointstrlist.Endpointstrs[k].Synstatelist = string(result)
				exist = true
				break
			}
		}
		if !exist {
			synstalist.Synstates = append(synstalist.Synstates, synstate)
			result, _ := json.Marshal(synstalist)
			epstr.Synstatelist = string(result)
			endpointstrlist.Endpointstrs = append(endpointstrlist.Endpointstrs, epstr)
		}
		result, _ := json.Marshal(endpointstrlist)
		synendpoint.Endpointstrlist = string(result)
		if err := synendpoint.Save(namespace, repository, tag); err != nil {
			return err
		}
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
		return err
	} else if resp.StatusCode != http.StatusOK {
		url := fmt.Sprintf("%s/syn/%s/%s/%s/recovery", dest, namespace, repository, tag)
		if recresp, err := module.SendHttpRequest("POST", url, nil, auth); err != nil {
			return err
		} else if recresp.StatusCode != http.StatusOK {
			return fmt.Errorf("response code %v", recresp.StatusCode)
		} else {
			return fmt.Errorf("response code %v", resp.StatusCode)
		}
	}

	if err := trigImageContent(namespace, repository, tag, auth, dest); err != nil {
		//synlog.Error("Failed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
		return err
	}

	return nil
}

func recoveryCont(namespace, repository, tag string) error {
	rec := new(Recovery)
	if exists, err := rec.Get(namespace, repository, tag); err != nil || !exists {
		return fmt.Errorf("no valid recovery data")
	}

	r := new(models.Repository)
	exists, err := r.Get(namespace, repository)
	if err != nil || !exists {
		return fmt.Errorf("invalid %s/%s", namespace, repository)
	}
	if err := json.Unmarshal([]byte(rec.Repobak), r); err != nil {
		return err
	}
	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	t := new(models.Tag)
	exists, err = t.Get(namespace, repository, tag)
	if err != nil || !exists {
		return fmt.Errorf("invalid %s/%s:%s", namespace, repository, tag)
	}
	if err := json.Unmarshal([]byte(rec.Tagbak), t); err != nil {
		return err
	}
	if err := t.Save(namespace, repository, tag); err != nil {
		return err
	}

	images := new([]models.Image)
	if err := json.Unmarshal([]byte(rec.Imagesbak), images); err != nil {
		return err
	}

	for _, imgbak := range *images {
		i := new(models.Image)
		*i = imgbak
	}

	if err := rec.Delete(namespace, repository, tag); err != nil {
		return err
	}

	//TODO: consider to recycle image content

	return nil
}

//TODO: must consider parallel, push/pull during synchron
func saveSynMetaDate(namespace, repository, tag string, reqbody []byte) error {
	sc := new(Syncont)
	sc.Layers = make(map[string][]byte)
	if err := json.Unmarshal(reqbody, sc); err != nil {
		return err
	}

	//local data recovery
	tb := new(models.Tag)
	if exists, err := tb.Get(namespace, repository, tag); err != nil {
		return err
	} else if exists {
		rec := new(Recovery)
		if _, err := rec.Get(namespace, repository, tag); err != nil {
			return err
		}

		if result, err := json.Marshal(tb); err != nil {
			return err
		} else {
			rec.Tagbak = string(result)
		}

		r := new(models.Repository)
		if exists, _ := r.Get(namespace, repository); !exists {
			return fmt.Errorf("not found repository %s/%s", namespace, repository)
		}
		if result, err := json.Marshal(r); err != nil {
			return err
		} else {
			rec.Repobak = string(result)
		}

		var images []models.Image
		tarsumlist, err := module.GetTarsumlist([]byte(tb.Manifest))
		if err != nil {
			return err
		}
		for _, tarsum := range tarsumlist {
			i := new(models.Image)
			if exists, _ := i.Get(tarsum); !exists {
				return fmt.Errorf("not found blob %s", tarsum)
			}
			images = append(images, *i)
		}
		result, err := json.Marshal(images)
		if err != nil {
			return err
		}
		rec.Imagesbak = string(result)

		if err := rec.Save(namespace, repository, tag); err != nil {
			return err
		}
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
	//if err := module.UploadLayer(tarsumlist); err != nil {
	//	return err
	//}

	return nil
}

func saveSynImgContent(ctx *macaron.Context, digest string) error {
	tarsum := strings.Split(digest, ":")[1]
	layerPath := module.GetLayerPath(tarsum, "layer", setting.APIVERSION_V2)

	file, err := os.Create(layerPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, ctx.Req.Request.Body); err != nil {
		return err
	}
	file.Close()
	var tarsumlist []string
	tarsumlist = append(tarsumlist, tarsum)

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
	}

	return nil
}

func trigImageContent(namespace, repository, tag, auth, dest string) error {
	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found tag %s/%s:%s", namespace, repository, tag)
	}

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

		//get layer from local or OSS
		if fd, err := module.DownloadLayer(i.Path); err != nil {
			return err
		} else {
			digest := "shas256:" + imageId
			url := fmt.Sprintf("%s/syn/%s/%s/%s/content/%s", dest, namespace, repository, tag, digest)
			if resp, err := module.SendHttpRequest("PUT", url, fd, auth); err != nil {
				//synlog.Error("Failed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
				return err
			} else if resp.StatusCode != http.StatusOK { //http.StatusInternalServerError
				return fmt.Errorf("response code %v", resp.StatusCode)
				//synlog.Error("Failed to synchronize %s/%s:%s to %s, err:%v", namespace, repository, tag, dest, err)
			}
		}
	}
	return nil
}

func saveRegionEndpoint(namespace, repository, tag string, reqbody []byte) error {
	eplist := new(Endpointlist)
	if err := json.Unmarshal(reqbody, eplist); err != nil {
		return err
	}

	regionIn := new(Region)
	if exists, err := regionIn.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exists {
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
					break
				}
			}

			if !exists {
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
		if err := updateRegionList(regionIn); err != nil {
			return err
		}
	}

	return nil
}

func saveEndpoint(regiontyp string, reqbody []byte) error {
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
					break
				}
			}

			if !exists {
				eplistOri.Endpoints = append(eplistOri.Endpoints, epin)
			}
		}
	} else {
		eplistOri = eplistIn
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

func updateRegionList(regionIn *Region) error {
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

func getSynList(regiontyp string) (string, error) {
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

func getRegionEndpoint(namespace, repository, tag string) (string, error) {
	r := new(Region)
	if exists, err := r.Get(namespace, repository, tag); err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	return r.Endpointlist, nil
}

func delRegionEndpoint(namespace, repository, tag string, reqbody []byte) (bool, error) {
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
		if err := updateRegionList(r); err != nil {
			return false, err
		}
	}

	return true, nil
}

func delEndpoint(regiontyp string, reqbody []byte) (bool, error) {
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
	return isStartup
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

		if err := saveSynMetaDate(namespace, repository, tag, body); err != nil {
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

func getSynchState(namespace, repository, tag string) (string, error) {
	se := new(SynEndpoint)
	if exists, err := se.Get(namespace, repository, tag); err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	return se.Endpointstrlist, nil
}
