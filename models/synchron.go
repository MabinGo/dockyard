package models

import (
	"fmt"

	"github.com/containerops/dockyard/utils/db"
)

type Region struct {
	Id         int64  `json:"id,omitempty" orm:"auto"`
	Namespace  string `json:"namespace,omitempty" orm:"null;varchar(255)"`
	Repository string `json:"repository,omitempty" orm:"null;varchar(255)"`
	Tag        string `json:"tag,omitempty" orm:"null;varchar(255)"`
	Region     string `json:"region" orm:"unique;null;varchar(255)"`
	Dest       string `json:"dest" orm:"unique;null;varchar(255)"` //TODO:要考虑支持组的概念
	Active     bool   `json:"active,omitempty" orm:"null;default(0)"`
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
	fmt.Println("####### Region.Save -1: ", exists, rgtmp)

	fmt.Println("####### Region.Save 0: ", exists, *rg)
	rg.Namespace, rg.Repository, rg.Tag = namespace, repository, tag
	if !exists {
		fmt.Println("####### Region.Save 1: ")
		err = db.Drv.Insert(rg)
	} else {
		fmt.Println("####### Region.Save 2: ", *rg)
		err = db.Drv.Update(rg)
	}

	return err
}
