package models

import (
	"github.com/containerops/dockyard/utils/db"
)

//distribution region
type Region struct {
	Id           int64  `json:"id,omitempty" orm:"auto"`
	Namespace    string `json:"namespace,omitempty" orm:"null;varchar(255)"`
	Repository   string `json:"repository,omitempty" orm:"null;varchar(255)"`
	Tag          string `json:"tag,omitempty" orm:"null;varchar(255)"`
	Endpointlist string `json:"endpointlist" orm:"null;type(text)"` //orm fk is invalid
}

type RegionTable struct {
	Id         int64  `json:"id,omitempty" orm:"auto"`
	Regionlist string `json:"regionlist" orm:"null;type(text)"`
}

type Endpoint struct {
	Area   string `json:"area"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Active bool   `json:"active,omitempty"`
}

//format of register region
type Endpointlist struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type Regionlist struct {
	Regions []Region `json:"region"`
}

type Syncont struct {
	Repository Repository        `json:"repository"`
	Tag        Tag               `json:"tag"`
	Images     []Image           `json:"images"`
	Layers     map[string][]byte `json:"layers"`
}

func (rt *RegionTable) Get(id int64) (bool, error) {
	rt.Id = id
	return db.Drv.Get(rt, id)
}

func (rt *RegionTable) Save(id int64) error {
	rttmp := RegionTable{Id: id}
	exists, err := rttmp.Get(id)
	if err != nil {
		return err
	}

	rt.Id = id
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
