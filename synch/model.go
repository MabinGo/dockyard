package synch

import (
	"fmt"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils/db"
)

//distribution region
type RegionTable struct {
	Id         int64  `json:"id" orm:"auto"`
	Name       string `json:"name" orm:"null;varchar(255)"`
	Regionlist string `json:"regionlist" orm:"null;type(text)"`
	DRClist    string `json:"drclist" orm:"null;type(text)"`
	Masterlist string `json:"masterlist" orm:"null;type(text)"`
}

type Region struct {
	Id           int64  `json:"id,omitempty" orm:"auto"`
	Namespace    string `json:"namespace,omitempty" orm:"null;varchar(255)"`
	Repository   string `json:"repository,omitempty" orm:"null;varchar(255)"`
	Tag          string `json:"tag,omitempty" orm:"null;varchar(255)"`
	Endpointlist string `json:"endpointlist" orm:"null;type(text)"` //orm fk is invalid
}

type Regionlist struct {
	Regions []Region `json:"region"`
}

//format of register region
type Endpointlist struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Area   string `json:"area"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
	Status string `json:"status"`
}

type Syncont struct {
	Repository models.Repository `json:"repository"`
	Tag        models.Tag        `json:"tag"`
	Images     []models.Image    `json:"images"`
	Layers     map[string][]byte `json:"layers"`
}

var (
	MASTER = "master"
	DRC    = "drc"
	COMMON = "common"
)

func (rt *RegionTable) Get(name string) (bool, error) {
	rt.Name = name
	return db.Drv.Get(rt, name)
}

func (rt *RegionTable) Save(name string) error {
	rttmp := RegionTable{Name: name}
	exists, err := rttmp.Get(name)
	if err != nil {
		return err
	}

	rt.Name = name
	if !exists {
		err = db.Drv.Insert(rt)
	} else {
		err = db.Drv.Update(rt)
	}

	return err
}

func (rg *Region) Get(namespace, repository, tag string) (bool, error) {
	rg.Namespace, rg.Repository, rg.Tag = namespace, repository, tag
	return db.Drv.Get(rg, namespace, repository, tag)
}

func (rg *Region) Save(namespace, repository, tag string) error {
	rgtmp := Region{Namespace: namespace, Repository: repository, Tag: tag}
	exists, err := rgtmp.Get(namespace, repository, tag)
	if err != nil {
		return err
	}

	rg.Namespace, rg.Repository, rg.Tag = namespace, repository, tag
	if !exists {
		err = db.Drv.Insert(rg)
	} else {
		err = db.Drv.Update(rg)
	}

	return err
}

func (rg *Region) Delete(namespace, repository, tag string) error {
	if exists, err := rg.Get(namespace, repository, tag); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("not found region")
	} else {
		return db.Drv.Delete(rg)
	}
}
