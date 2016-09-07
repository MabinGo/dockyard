/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/containerops/dockyard/db"
)

type AppV1 struct {
	Id          int64      `json:"id" gorm:"primary_key"`
	Namespace   string     `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository  string     `json:"repository" sql:"not null;type:varchar(255)"`
	Path        string     `json:"path"`
	Description string     `json:"description" sql:"null;type:text"`
	Manifests   string     `json:"manifests" sql:"null;type:text"`
	Keys        string     `json:"-" sql:"null;type:text"`
	Size        int64      `json:"size" sql:"default:0"`
	NumPackages int64      `json:"num_packages" sql:"-"`
	Locked      int64      `json:"-" sql:"default:false"`
	CreatedAt   time.Time  `json:"created" sql:""`
	UpdatedAt   time.Time  `json:"updated" sql:""`
	DeletedAt   *time.Time `json:"deleted" sql:"index"`
}

type ArtifactV1 struct {
	Id        int64      `json:"id" gorm:"primary_key"`
	AppV1     int64      `json:"appv1" sql:"not null"`
	OS        string     `json:"os" sql:"null;type:varchar(128)"`
	Arch      string     `json:"arch" sql:"null;type:varchar(128)"`
	Type      string     `json:"type" sql:"null;type:varchar(64)"`
	App       string     `json:"app" sql:"not null;varchar(255)"`
	Tag       string     `json:"tag" sql:"not null;varchar(255)"`
	BlobSum   string     `json:"blobsum" sql:"type:varchar(255)"`
	OSS       string     `json:"oss" sql:"null;type:text"`
	Manifests string     `json:"manifests" sql:"null;type:text"`
	URL       string     `json:"url" sql:"null;type:text"`
	Path      string     `json:"path" sql:"null;type:text"`
	Size      int64      `json:"size" sql:"default:0"`
	Locked    int64      `json:"locked" sql:"default:0"` // If the sql field name is changed, update the FreeLock method!
	CreatedAt time.Time  `json:"created" sql:""`
	UpdatedAt time.Time  `json:"updated" sql:""`
	DeletedAt *time.Time `json:"deleted" sql:"index"`
}

type SearchOutput struct {
	Namespace   string    `json:"namespace" description:"application's namespace"`
	Repository  string    `json:"repository" description:"name of application's repository"`
	OS          string    `json:"os" description:"os type of application, default is 'undefine'"`
	Arch        string    `json:"arch" description:"architecture of application, default is 'undefine'"`
	Name        string    `json:"name" description:"application's name"`
	Tag         string    `json:"tag" description:"application's tag"`
	Description string    `json:"description" description:"application's description"`
	URL         string    `json:"url" description:"application's downloading url"`
	Size        int64     `json:"size" description:"application's size"`
	CreatedAt   time.Time `json:"createdat" description:"application's created time"`
	UpdatedAt   time.Time `json:"updatedat" description:"application's updated time"`
}

func (a *AppV1) TableName() string {
	return "app_v1"
}

func (a *AppV1) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(a, "idx_appv1_namespace_repository",
		"namespace", "repository"); err != nil {
		return fmt.Errorf("create unique index idx_appv1_namespace_repository error:" + err.Error())
	}
	return nil
}

func (a *AppV1) IsExist() (bool, error) {
	if records, err := db.Instance.Count(a); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (a *AppV1) Save(condition *AppV1) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		err = db.Instance.Create(a)
	} else {
		a.Id = condition.Id
		err = db.Instance.Update(a)
	}

	return err
}

func (a *AppV1) Delete() error {
	exists, err := a.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("record not found")
	} else {
		err = db.Instance.Delete(a)
	}

	return err
}

func (a *AppV1) List(results *[]AppV1) error {
	if err := db.Instance.QueryM(a, results); err != nil {
		if strings.EqualFold(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	return nil
}

func (*ArtifactV1) TableName() string {
	return "artifact_v1"
}

func (i *ArtifactV1) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(i, "idx_artifactv1_appv1id_os_arch_app_tag_type",
		"app_v1", "os", "arch", "app", "tag", "type"); err != nil {
		return fmt.Errorf("create unique index idx_artifactv1_appv1id_name_os_arch_tag_type error:" + err.Error())
	}

	if err := db.Instance.AddForeignKey(i, "app_v1", "app_v1(id)", "CASCADE", "NO ACTION"); err != nil {
		return fmt.Errorf("create foreign key app_v1 error:" + err.Error())
	}
	return nil
}

func (i *ArtifactV1) IsExist() (bool, error) {
	if records, err := db.Instance.Count(i); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, err
	}
	return false, nil
}

func (i *ArtifactV1) Read() (bool, error) {
	if records, err := db.Instance.Count(i); err != nil {
		return false, err
	} else if records > int64(0) {
		if i.Locked == 0 || i.Locked == 1 {
			i.Locked = 1
			db.Instance.Update(i)
			return true, err
		}
		return false, fmt.Errorf("source is busy")
	}
	return false, nil
}

func (i *ArtifactV1) Create() error {
	return db.Instance.Create(i)
}

func (i *ArtifactV1) Update() error {
	return db.Instance.Save(i)
}

func (i *ArtifactV1) Save(condition *ArtifactV1) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if condition.Locked != 0 {
		return fmt.Errorf("source is busy")
	}

	i.Locked = 2
	if !exists {
		err = db.Instance.Create(i)
	} else {
		i.Id = condition.Id
		err = db.Instance.Update(i)
	}

	return err
}

func (i *ArtifactV1) SaveAtom(condition *ArtifactV1) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if condition.Locked != 0 {
		return fmt.Errorf("source is busy")
	}

	if !exists {
		err = db.Instance.Create(i)
	} else {
		i.Id = condition.Id
		err = db.Instance.Update(i)
	}

	return err
}

func (i *ArtifactV1) UpdateBlob(blobsum string) (string, error) {
	digest := ""

	i.Locked = 0
	// update blobsum and return redundancy blobsum
	if i.BlobSum != "" {
		if strings.Compare(i.BlobSum, blobsum) != 0 {
			records, err := db.Instance.Count(&ArtifactV1{BlobSum: blobsum})
			if err != nil {
				return "", err
			}
			if records == 1 {
				digest = blobsum
			}
		}
	}
	err := db.Instance.Save(i)

	return digest, err
}

func (i *ArtifactV1) Delete() (string, error) {
	digest := ""
	exists, err := i.IsExist()
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("record not found")
	} else {
		if i.Locked != 0 {
			return "", fmt.Errorf("Source is busy")
		}
		records, err := db.Instance.Count(&ArtifactV1{BlobSum: i.BlobSum})
		if err != nil {
			return "", err
		}
		if records == 1 {
			digest = i.BlobSum
		}
		err = db.Instance.Delete(i)
	}

	return digest, err
}

//DeleteM is the helper to delete multiple records by the given condition
func (i *ArtifactV1) DeleteM(query string) error {
	return db.Instance.BatchDelete(&i, query)
}

func (i *ArtifactV1) List(results *[]ArtifactV1) error {
	if err := db.Instance.QueryM(i, results); err != nil {
		if strings.EqualFold(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	return nil
}

func (i *ArtifactV1) Count(countField string, condition string) (int64, error) {
	sql := fmt.Sprintf("SELECT COUNT(%s) FROM artifact_v1 %s", countField, condition)
	var count int64
	db.Instance.Exec(sql).Row().Scan(&count)
	return count, nil
}

//QueryGlobal fuzzy query
func (i *ArtifactV1) QueryGlobal(results interface{}, parameters ...string) error {
	sql := "select * from artifact_v1 where deleted_at is null and "
	concat := "concat(app, \", \", os, \", \", arch, \", \", tag) like ?"
	switch len(parameters) {
	case 0:
		return fmt.Errorf("invalid query parameters")
	case 1:
		sql = sql + concat
		db.Instance.Raw(results, sql, "%"+parameters[0]+"%")
	case 2:
		sql = sql + concat + " and " + concat
		db.Instance.Raw(results, sql, "%"+parameters[0]+"%", "%"+parameters[1]+"%")
	case 3:
		sql = sql + concat + " and " + concat + " and " + concat
		db.Instance.Raw(results, sql, "%"+parameters[0]+"%", "%"+parameters[1]+"%", "%"+parameters[2]+"%")
	case 4:
		sql = sql + concat + " and " + concat + " and " + concat + " and " + concat
		db.Instance.Raw(results, sql, "%"+parameters[0]+"%", "%"+parameters[1]+"%", "%"+parameters[2]+"%", "%"+parameters[3]+"%")
	default:
	}

	return nil
}

func (i *ArtifactV1) QueryScope(results interface{}, parameters ...string) error {
	appv1 := i.AppV1
	sql := "select * from artifact_v1 where deleted_at is null and app_v1=? and "
	concat := "concat(app, \", \", os, \", \", arch, \", \", tag) like ?"
	switch len(parameters) {
	case 0:
		return fmt.Errorf("invalid query parameters")
	case 1:
		sql = sql + concat
		db.Instance.Raw(results, sql, appv1, "%"+parameters[0]+"%")
	case 2:
		sql = sql + concat + " and " + concat
		db.Instance.Raw(results, sql, appv1, "%"+parameters[0]+"%", "%"+parameters[1]+"%")
	case 3:
		sql = sql + concat + " and " + concat + " and " + concat
		db.Instance.Raw(results, sql, appv1, "%"+parameters[0]+"%", "%"+parameters[1]+"%", "%"+parameters[2]+"%")
	case 4:
		sql = sql + concat + " and " + concat + " and " + concat + " and " + concat
		db.Instance.Raw(results, sql, appv1, "%"+parameters[0]+"%", "%"+parameters[1]+"%", "%"+parameters[2]+"%", "%"+parameters[3]+"%")
	}

	return nil
}

func (i *ArtifactV1) FreeLock() error {
	// i.Locked = 0
	// return db.Instance.Update(i)
	return db.Instance.UpdateField(i, "locked", 0)
}

type State struct {
	Id         int64     `json:"id" gorm:"primary_key"`
	Namespace  string    `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository string    `json:"repository" sql:"not null;type:varchar(255)"`
	UUID       string    `json:"uuid" sql:"not null;type:varchar(128)"`
	Offset     int64     `json:"offset" sql:"default:0"`
	Locked     int64     `json:"-" sql:"default:false"`
	CreatedAt  time.Time `json:"created" sql:""`
	//Action     string    `json:"action" sql:"not null;type:varchar(32)"`
}

func (s *State) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(s, "idx_state_namespace_repository", "namespace", "repository"); err != nil {
		return fmt.Errorf("create unique index idx_state_namespace_repository error:" + err.Error())
	}
	return nil
}

func (s *State) IsExist() (bool, error) {
	if records, err := db.Instance.Count(s); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (s *State) Save(condition *State) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		err = db.Instance.Create(s)
	} else {
		s.Id = condition.Id
		err = db.Instance.Update(s)
	}

	return err
}
