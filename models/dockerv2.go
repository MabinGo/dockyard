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

type DockerV2 struct {
	Id            int64      `json:"id" gorm:"primary_key"`
	Namespace     string     `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository    string     `json:"repository" sql:"not null;type:varchar(255)"`
	SchemaVersion string     `json:"schemaversion" sql:"not null;type:varchar(255)"`
	Manifests     string     `json:"manifests" sql:"null;type:mediumtext"`
	Agent         string     `json:"agent" sql:"null;type:text"`
	Description   string     `json:"description" sql:"null;type:text"`
	Size          int64      `json:"size" sql:"default:0"`
	Path          string     `json:"path" sql:"-"`
	Locked        int64      `json:"locked" sql:"default:0"`
	IsPublic      bool       `json:"is_public" sql:"default:false"`
	NumImages     int64      `json:"num_images" sql:"-"`
	Tags          []string   `json:"tags,omitempty" sql:"-"`
	CreatedAt     time.Time  `json:"created" sql:""`
	UpdatedAt     time.Time  `json:"updated" sql:""`
	DeletedAt     *time.Time `json:"deleted" sql:"index"`
}

type DockerV2Resp struct {
	Namespace     string    `json:"namespace" description:"The namespace that the docker image belongs to"`
	Repository    string    `json:"repository" description:"The docker image's name"`
	SchemaVersion string    `json:"schemaversion" description:"The schema version of the image"`
	Manifests     string    `json:"manifests" description:"The content of the image's manifests file"`
	Size          int64     `json:"size" description:"The size of the image, which is the sum of all tags' size"`
	Path          string    `json:"path" description:"The path for docker pull command"`
	IsPublic      bool      `json:"is_public" description:"Is the docker image public or not"`
	NumImages     int64     `json:"num_images" description:"The number of the image's tags"`
	Tags          []string  `json:"tags,omitempty" description:"The array of the image's tags"`
	CreatedAt     time.Time `json:"created" description:"Time when the image was created"`
	UpdatedAt     time.Time `json:"updated" description:"The last time when the image was updated"`
}

type DockerTagV2 struct {
	Id        int64      `json:"id" gorm:"primary_key"`
	DockerV2  int64      `json:"dockerv2" sql:"not null"`
	Tag       string     `json:"tag" sql:"not null;type:varchar(255)"`
	ImageId   string     `json:"imageid" sql:"not null;type:varchar(255)"`
	Manifest  string     `json:"manifest" sql:"null;type:text"`
	Schema    int64      `json:"schema" sql:""`
	Locked    int64      `json:"locked" sql:"default:0"`
	Path      string     `json:"path" sql:"-"`
	Size      int64      `json:"size" sql:"-"`
	CreatedAt time.Time  `json:"created" sql:""`
	UpdatedAt time.Time  `json:"updated" sql:""`
	DeletedAt *time.Time `json:"deleted" sql:"index"`
}

type DockerTagV2Resp struct {
	Tag       string    `json:"tag" description:"Tag of the image"`
	ImageId   string    `json:"imageid" description:"Id of the image"`
	Manifest  string    `json:"manifest" description:"The image's manifest file content"`
	Schema    int64     `json:"schema" description:"Schema of the image"`
	Path      string    `json:"path" description:"The path for docker pull command"`
	Size      int64     `json:"size" description:"The sum of the layer's and manifest file's size"`
	CreatedAt time.Time `json:"created" description:"The time when the image(with tag) was created"`
	UpdatedAt time.Time `json:"updated" description:"The last time when the image(with tag) was updated"`
}

type DockerImageV2 struct {
	Id              int64      `json:"id" gorm:"primary_key"`
	ImageId         string     `json:"imageid" sql:"null;type:varchar(255)"`
	BlobSum         string     `json:"blobsum" sql:"unique;type:varchar(255)"`
	V1Compatibility string     `json:"v1compatibility" sql:"null;type:text"`
	Path            string     `json:"path" sql:"null;type:text"`
	OSS             string     `json:"oss" sql:"null;type:text"`
	Size            int64      `json:"size" sql:"default:0"`
	Reference       int64      `json:"reference" sql:"default:0"`
	Locked          int64      `json:"locked" sql:"default:0"`
	CreatedAt       time.Time  `json:"created" sql:""`
	UpdatedAt       time.Time  `json:"updated" sql:""`
	DeletedAt       *time.Time `json:"deleted" sql:"index"`
}

func (*DockerV2) TableName() string {
	return "docker_V2"
}

func (r *DockerV2) Resp() *DockerV2Resp {
	return &DockerV2Resp{
		Namespace:     r.Namespace,
		Repository:    r.Repository,
		SchemaVersion: r.SchemaVersion,
		Manifests:     r.Manifests,
		Size:          r.Size,
		Path:          r.Path,
		IsPublic:      r.IsPublic,
		NumImages:     r.NumImages,
		Tags:          r.Tags,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func (r *DockerV2) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(r, "idx_dockerv2_namespace_repository",
		"namespace", "repository"); err != nil {
		return fmt.Errorf("create unique index idx_dockerv2_namespace_repository error:" + err.Error())
	}
	return nil
}

func (r *DockerV2) IsExist() (bool, error) {
	if records, err := db.Instance.Count(r); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (r *DockerV2) Save(condition *DockerV2) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		err = db.Instance.Create(r)
	} else {
		r.Id = condition.Id
		err = db.Instance.Update(r)
	}

	return err
}

func (r *DockerV2) Update() error {
	if err := db.Instance.Save(r); err != nil {
		return err
	}

	return nil
}

func (r *DockerV2) Delete() error {
	exists, err := r.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("record not found")
	} else {
		err = db.Instance.Delete(r)
	}

	return err
}

func (r *DockerV2) List(results *[]DockerV2) error {
	if err := db.Instance.QueryM(r, results); err != nil {
		if strings.EqualFold(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	return nil
}

func (*DockerImageV2) TableName() string {
	return "docker_image_v2"
}

func (i *DockerImageV2) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(i, "idx_dockerimagev2_blobsum",
		"blob_sum"); err != nil {
		return fmt.Errorf("create unique index idx_dockerimagev2_blobsum error:" + err.Error())
	}
	return nil
}

func (i *DockerImageV2) IsExist() (bool, error) {
	if records, err := db.Instance.Count(i); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (i *DockerImageV2) Read() (bool, error) {
	if records, err := db.Instance.Count(i); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (i *DockerImageV2) Write() (bool, error) {
	if records, err := db.Instance.Count(i); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (i *DockerImageV2) Save() error {
	exists, err := i.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		err = db.Instance.Create(i)
	} else {
		err = db.Instance.Update(i)
	}

	return err
}

func (i *DockerImageV2) SaveAtom(condition *DockerImageV2) error {
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

func (i *DockerImageV2) Update() error {
	if err := db.Instance.Save(i); err != nil {
		return err
	}

	return nil
}

func (i *DockerImageV2) Delete() error {
	exists, err := i.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("record not found")
	} else {
		err = db.Instance.Delete(i)
	}

	return err
}

func (i *DockerImageV2) FreeLock() error {
	i.Locked = 0
	return db.Instance.Save(i)
}

func (*DockerTagV2) TableName() string {
	return "docker_tag_V2"
}

func (t *DockerTagV2) Resp() *DockerTagV2Resp {
	return &DockerTagV2Resp{
		Tag:       t.Tag,
		ImageId:   t.ImageId,
		Manifest:  t.Manifest,
		Schema:    t.Schema,
		Path:      t.Path,
		Size:      t.Size,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (t *DockerTagV2) AddUniqueIndex() error {
	if err := db.Instance.AddUniqueIndex(t, "idx_dockertagv2_dockerv2_tag",
		"docker_v2", "tag"); err != nil {
		return fmt.Errorf("create unique index idx_dockertagv2_dockerv2_tag error:" + err.Error())
	}
	return nil
}

func (t *DockerTagV2) IsExist() (bool, error) {
	if records, err := db.Instance.Count(t); err != nil {
		return false, err
	} else if records > int64(0) {
		return true, nil
	}
	return false, nil
}

func (t *DockerTagV2) Save(condition *DockerTagV2) error {
	exists, err := condition.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		err = db.Instance.Create(t)
	} else {
		t.Id = condition.Id
		err = db.Instance.Update(t)
	}

	return err
}

func (t *DockerTagV2) Delete() error {
	exists, err := t.IsExist()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("record not found")
	} else {
		err = db.Instance.Delete(t)
	}

	return err
}

func (t *DockerTagV2) Count(countField string, condition string) (int64, error) {
	sql := fmt.Sprintf("SELECT COUNT(%s) FROM %s %s", countField, t.TableName(), condition)
	var count int64
	db.Instance.Exec(sql).Row().Scan(&count)
	return count, nil
}

func (t *DockerTagV2) List(results *[]DockerTagV2) error {
	if err := db.Instance.QueryM(t, results); err != nil {
		if strings.EqualFold(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	return nil
}

type Repolist struct {
	Repositories []string `json:"repositories" description:"repositories list"`
}

type Taglist struct {
	Name string   `json:"name" description:"namespace/repository"`
	Tags []string `json:"tags" description:"all the tags in namespace/repository"`
}
