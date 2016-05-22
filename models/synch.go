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

type Endpoint struct {
	Area   string `json:"area"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Active bool   `json:"active,omitempty"`
}

//format of register region
type Endpointgrp struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type Syncont struct {
	Repository Repository        `json:"repository"`
	Tag        Tag               `json:"tag"`
	Images     []Image           `json:"images"`
	Layers     map[string][]byte `json:"layers"`
}

var Regions = []Region{}
var SynConts = []Syncont{}

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
