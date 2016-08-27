package synch

import (
	"fmt"
	"time"

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

type SynEndpoint struct {
	Id              int64  `json:"id,omitempty" orm:"auto"`
	Namespace       string `json:"namespace,omitempty" orm:"null;varchar(255)"`
	Repository      string `json:"repository,omitempty" orm:"null;varchar(255)"`
	Tag             string `json:"tag,omitempty" orm:"null;varchar(255)"`
	Endpointstrlist string `json:"endpointstr,omitempty" orm:"null;type(text)"`
}

type Recovery struct {
	Id         int64  `json:"id,omitempty" orm:"auto"`
	Namespace  string `json:"namespace" orm:"null;varchar(255)"`
	Repository string `json:"repository" orm:"null;varchar(255)"`
	Tag        string `json:"tag" orm:"null;varchar(255)"`
	Repobak    string `json:"repobak" orm:"null;type(text)"`
	Tagbak     string `json:"tagbak" orm:"null;type(text)"`
	Imagesbak  string `json:"imagesbak" orm:"null;type(text)"`
	//Layers     map[string][]byte `json:"layers"`
}

func (rec *Recovery) TableUnique() [][]string {
	return [][]string{
		{"Namespace", "Repository", "Tag"},
	}
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

func (r *SynEndpoint) TableUnique() [][]string {
	return [][]string{
		{"Namespace", "Repository", "Tag"},
	}
}

type Endpointstrlist struct {
	Endpointstrs []Endpointstr `json:"endpointstrss"`
}

type Endpointstr struct {
	Area         string `json:"area"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Synstatelist string `json:"synstatelist" orm:"null;type(text)"`
}

type Synstatelist struct {
	Synstates []Synstate `json:"synstates"`
}

type Synstate struct {
	Status     string    `json:"status"`
	Statuscode string    `json:"statuscode"`
	Response   string    `json:"response"`
	Time       time.Time `json:"time"`
}

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

func (se *SynEndpoint) Get(namespace, repository, tag string) (bool, error) {
	se.Namespace, se.Repository, se.Tag = namespace, repository, tag
	return db.Drv.Get(se, namespace, repository, tag)
}

func (se *SynEndpoint) Save(namespace, repository, tag string) error {
	setmp := SynEndpoint{Namespace: namespace, Repository: repository, Tag: tag}
	exists, err := setmp.Get(namespace, repository, tag)
	if err != nil {
		return err
	}

	se.Namespace, se.Repository, se.Tag = namespace, repository, tag
	if !exists {
		err = db.Drv.Insert(se)
	} else {
		err = db.Drv.Update(se)
	}

	return err
}

func (rec *Recovery) Get(namespace, repository, tag string) (bool, error) {
	rec.Namespace, rec.Repository, rec.Tag = namespace, repository, tag
	return db.Drv.Get(rec, namespace, repository, tag)
}

func (rec *Recovery) Save(namespace, repository, tag string) error {
	rectmp := Recovery{Namespace: namespace, Repository: repository, Tag: tag}
	exists, err := rectmp.Get(namespace, repository, tag)
	if err != nil {
		return err
	}

	rec.Namespace, rec.Repository, rec.Tag = namespace, repository, tag
	if !exists {
		err = db.Drv.Insert(rec)
	} else {
		err = db.Drv.Update(rec)
	}

	return err
}

func (rec *Recovery) Delete(namespace, repository, tag string) error {
	recovery := Recovery{Namespace: namespace, Repository: repository, Tag: tag}
	_, err := recovery.Get(namespace, repository, tag)
	if err != nil {
		return err
	}

	rec.Namespace, rec.Repository, rec.Tag = namespace, repository, tag
	err = db.Drv.Delete(rec)

	return err

}
