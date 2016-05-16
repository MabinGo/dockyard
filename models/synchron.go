package models

import (
	"github.com/containerops/dockyard/utils/db"
)

type Region struct {
	Id         int64  `json:"id,omitempty" orm:"auto"`
	Namespace  string `json:"namespace,omitempty" orm:"varchar(255)"`
	Repository string `json:"repository,omitempty" orm:"varchar(255)"`
	Tag        string `json:"tag,omitempty" orm:"varchar(255)"`
	Region     string `json:"region" orm:"unique;varchar(255)"`
	Dest       string `json:"dest" orm:"unique;varchar(255)"` //TODO:要考虑支持组的概念
	Active     bool   `json:"active,omitempty" orm:"null;default(0)"`
}

type Syn struct {
	Repository Repository
	Tag        Tag
	Images     []Image
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
