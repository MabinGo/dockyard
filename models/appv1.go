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
	"net/url"
	"strings"
	"time"

	"github.com/containerops/dockyard/db"
)

type AppV1 struct {
	Id          int64      `json:"id" gorm:"primary_key"`
	Namespace   string     `json:"namespace" sql:"null;type:varchar(255)"`
	Repository  string     `json:"repository" sql:"null;type:varchar(255)"`
	Path        string     `json:"path"`
	Description string     `json:"description" sql:"null;type:text"`
	Manifests   string     `json:"manifests" sql:"null;type:mediumtext"`
	Keys        string     `json:"-" sql:"null;type:text"`
	Size        int64      `json:"size" sql:"default:0"`
	NumPackages int64      `json:"num_packages" sql:"-"`
	IsPublic    bool       `json:"is_public" sql:"default:false"`
	Locked      int64      `json:"-" sql:"default:false"`
	CreatedAt   time.Time  `json:"created" sql:""`
	UpdatedAt   time.Time  `json:"updated" sql:""`
	DeletedAt   *time.Time `json:"deleted" sql:"index"`
}

type AppV1Resp struct {
	Namespace   string    `json:"namespace" description:"The namespace that the repository belongs to"`
	Repository  string    `json:"repository" description:"The repository name"`
	Path        string    `json:"path" description:"NAMESPACE/REPOSITORY"`
	Manifests   string    `json:"manifests" description:"The descriptive text of the repository"`
	Size        int64     `json:"size" description:"The sum of the repository's packages size"`
	NumPackages int64     `json:"num_packages" description:"The number of the repository's packages"`
	IsPublic    bool      `json:"is_public" description:"Whether the repository is public"`
	CreatedAt   time.Time `json:"created" description:"The time when the repository was created"`
	UpdatedAt   time.Time `json:"updated" description:"The last time when the repository was updated"`
}

type ArtifactV1 struct {
	Id        int64      `json:"id" gorm:"primary_key"`
	AppV1     int64      `json:"appv1" sql:"null"`
	OS        string     `json:"os" sql:"null;type:varchar(128)"`
	Arch      string     `json:"arch" sql:"null;type:varchar(128)"`
	App       string     `json:"app" sql:"null;varchar(255)"`
	Tag       string     `json:"tag" sql:"null;varchar(255)"`
	BlobSum   string     `json:"blobsum" sql:"type:varchar(255)"`
	OSS       string     `json:"oss" sql:"null;type:text"`
	Manifests string     `json:"manifests" sql:"null;type:text"`
	URL       string     `json:"url" sql:"null;type:text"`
	Path      string     `json:"path" sql:"null;type:text"`
	Size      int64      `json:"size" sql:"default:0"`
	Locked    int64      `json:"locked" sql:"default:0"` // If the sql field name is changed, update the FreeLock method!
	Active    int64      `json:"active" sql:"default:0"`
	CreatedAt time.Time  `json:"created" sql:""`
	UpdatedAt time.Time  `json:"updated" sql:""`
	DeletedAt *time.Time `json:"deleted" sql:"index"`
}

type ArtifactV1Resp struct {
	Namespace  string    `json:"namespace" description:"The namespace that the package belongs to"`
	Repository string    `json:"repository" description:"The repository that the package belongs to"`
	OS         string    `json:"os" description:"Operating system"`
	Arch       string    `json:"arch" description:"Platform architecture"`
	App        string    `json:"app" description:"Package's name"`
	Tag        string    `json:"tag" description:"Package's version"`
	Manifests  string    `json:"manifests" description:"A descriptive text of the package"`
	URL        string    `json:"url" description:"The package's file download url"`
	WebURL     string    `json:"web_url" description:"The package's file download url(web/v1)"`
	Size       int64     `json:"size" description:"The file size of the package"`
	CreatedAt  time.Time `json:"created" description:"Time when package was created"`
	UpdatedAt  time.Time `json:"updated" description:"The last time when package was updated"`
}

type SearchOutput struct {
	Namespace   string    `json:"namespace" description:"application's namespace"`
	Repository  string    `json:"repository" description:"name of application's repository"`
	OS          string    `json:"os" description:"os type of application, default is 'undefined'"`
	Arch        string    `json:"arch" description:"architecture of application, default is 'undefined'"`
	Name        string    `json:"name" description:"application's name"`
	Tag         string    `json:"tag" description:"application's tag"`
	Description string    `json:"description" description:"application's description"`
	URL         string    `json:"url" description:"application's downloading url"`
	Size        int64     `json:"size" description:"application's size"`
	CreatedAt   time.Time `json:"createdat" description:"application's created time"`
	UpdatedAt   time.Time `json:"updatedat" description:"application's updated time"`
}

type Session struct {
	Id         int64     `json:"id" gorm:"primary_key"`
	Namespace  string    `json:"namespace" sql:"null;type:varchar(255)"`
	Repository string    `json:"repository" sql:"null;type:varchar(255)"`
	Imageid    int64     `json:"imageid" sql:"null"`
	Version    int64     `json:"version" sql:"default:0"`
	UUID       string    `json:"uuid" sql:"null;type:varchar(128)"`
	Locked     int64     `json:"locked" sql:"default:0"`
	CreatedAt  time.Time `json:"created_at" sql:""`
	UpdatedAt  time.Time `json:"updated_at" sql:""`
	//Operator   string    `json:"operator" sql:"type:varchar(255)"`
	//Action string `json:"action" sql:"not null;type:varchar(32)"`
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

func (a *AppV1) Resp() *AppV1Resp {
	return &AppV1Resp{
		Namespace:   a.Namespace,
		Repository:  a.Repository,
		Path:        a.Path,
		Manifests:   a.Manifests,
		Size:        a.Size,
		NumPackages: a.NumPackages,
		IsPublic:    a.IsPublic,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

//
func (a *AppV1) Update() error {
	return db.Instance.Save(a)
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

func (i *ArtifactV1) TableName() string {
	return "artifact_v1"
}

func (i *ArtifactV1) Resp(u *url.URL, namespace, repository string) *ArtifactV1Resp {
	resp := &ArtifactV1Resp{
		OS:        i.OS,
		Arch:      i.Arch,
		App:       i.App,
		Tag:       i.Tag,
		Manifests: i.Manifests,
		Size:      i.Size,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
	resp.Namespace = namespace
	resp.Repository = repository

	resp.URL = fmt.Sprintf("%s://%s/app/v1/%s/%s/%s/%s/%s/%s", u.Scheme, u.Host, namespace, repository, i.OS, i.Arch, i.App, i.Tag)
	resp.WebURL = fmt.Sprintf("%s://%s/web/v1/app/%s/%s/%s/%s/%s/%s/file/%s", u.Scheme, u.Host, namespace, repository, i.OS, i.Arch, i.App, i.Tag, i.App)
	return resp
}

func (i *ArtifactV1) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(i, "idx_artifactv1_appv1id_os_arch_app_tag",
		"app_v1", "os", "arch", "app", "tag"); err != nil {
		return fmt.Errorf("create unique index idx_artifactv1_appv1id_name_os_arch_tag error:" + err.Error())
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
		return true, err
	}
	return false, nil
}

func (i *ArtifactV1) Create() error {
	return db.Instance.Create(i)
}

func (i *ArtifactV1) Update() error {
	return db.Instance.Save(i)
}

func (i *ArtifactV1) UpdateImageStatus(value int64) error {
	return db.Instance.UpdateField(i, "active", value)
}

func (i *ArtifactV1) Save(condition *ArtifactV1) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

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

func (s *Session) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(s, "idx_session_namespace_repository_version", "namespace", "repository", "version"); err != nil {
		return fmt.Errorf("create unique index idx_session_namespace_repository_version error:" + err.Error())
	}
	return nil
}

func (s *Session) isExist() (bool, error) {
	if records, err := db.Instance.Count(s); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (s *Session) Save(namespace, repository string, version int64) error {
	condition := new(Session)
	condition.Namespace, condition.Repository, condition.Version = namespace, repository, version
	exists, err := condition.isExist()
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

func (s *Session) Read(namespace, repository string, version int64) (bool, error) {
	s.Namespace, s.Repository, s.Version = namespace, repository, version
	if records, err := db.Instance.Count(s); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (s *Session) Delete() error {
	exists, err := s.isExist()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("record not found")
	} else {
		err = db.Instance.Delete(s)
	}

	return err
}

func (s *Session) UpdateSessionLock(value int64) error {
	return db.Instance.UpdateField(s, "locked", value)
}

func (s *Session) TableLock() error {
	t := new(Session)

	cmd := "SET AUTOCOMMIT=0"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = fmt.Sprintf("LOCK TABLES %s WRITE", t.TableName())
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func (s *Session) TableUnlock() error {
	cmd := "COMMIT"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}

	cmd = "UNLOCK TABLES"
	if err := db.Instance.Exe(cmd); err != nil {
		return err
	}
	return nil
}

func (s *Session) TableName() string {
	return "sessions"
}

func (s *Session) Find(results *[]Session) (int64, error) {
	count, err := db.Instance.Find(results)
	if err != nil {
		return 0, err
	}

	return count, nil
}
