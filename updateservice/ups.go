/*
Copyright 2016 The ContainerOps Authors All rights reserved.

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

package us

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
)

const (
	defaultMetaDirName      = "meta"
	defaultMetaFileName     = "meta.json"
	defaultMetaSignFileName = "meta.sign"

	//The default life circle for a software is half a year
	defaultLifecircle = time.Hour * 24 * 180
)

var (
	metaMutex sync.Mutex
)

// UpdateService is the open struct to sign/encrypt a file
type UpdateService struct {
}

// Put adds a file to the update service system
func (us UpdateService) Put(proto, namespace, repository, fullname string, shas []string) error {
	meta, err := NewUpdateServiceMeta(proto, namespace, repository)
	if err != nil {
		return err
	}

	item, err := NewUpdateServiceItem(fullname, shas)
	if err != nil {
		return err
	}

	return meta.Put(item)
}

// Delete removes a file from the update service system
func (us UpdateService) Delete(proto, namespace, repository, fullname string) error {
	meta, err := NewUpdateServiceMeta(proto, namespace, repository)
	if err != nil {
		return err
	}

	return meta.Delete(fullname)
}

// ReadPublicKey reads the public key
func (us UpdateService) ReadPublicKey(proto, namespace string) ([]byte, error) {
	err := KeyManagerEnabled()
	if err != nil {
		return nil, err
	}

	mode, _ := NewKeyManagerMode(setting.KeyManagerMode)
	manager, _ := NewKeyManager(mode, setting.KeyManagerURI)

	return manager.ReadPublicKey(proto, namespace)
}

// ReadMeta reads the meta data
func (us UpdateService) ReadMeta(proto, namespace, repository string) ([]byte, error) {
	meta, err := NewUpdateServiceMeta(proto, namespace, repository)
	if err != nil {
		return nil, err
	}

	return meta.ReadMeta()
}

// ReadMetaSign reads the meta sign data
func (us UpdateService) ReadMetaSign(proto, namespace, repository string) ([]byte, error) {
	err := KeyManagerEnabled()
	if err != nil {
		return nil, err
	}

	meta, err := NewUpdateServiceMeta(proto, namespace, repository)
	if err != nil {
		return nil, err
	}

	return meta.ReadMetaSign()
}

// UpdateServiceItem keeps the meta data of a vm/app/image
type UpdateServiceItem struct {
	// Full represents a uniq name of a file within a repo, for app, fullname means os/arch/appname/tag
	FullName string
	// SHAS represents a sha list of a file.
	// If a file is composed of several layers or parts, len of SHAS will be bigger than one
	SHAS []string
	// Created is the created data of a vm/app/image
	Created time.Time
	// Updated is the latest updated data of a vm/app/image
	Updated time.Time
	// Expired is used to check if a vm/app/image need to be upgraded
	Expired time.Time
}

// NewUpdateServiceItem creates a service item by a 'FullName' and a 'SHA' list
func NewUpdateServiceItem(fn string, shas []string) (usi UpdateServiceItem, err error) {
	usi.FullName, usi.SHAS = fn, shas
	usi.Created = time.Now()
	usi.Updated = usi.Created
	usi.Expired = usi.Created.Add(defaultLifecircle)

	if ok, err := usi.isValid(); !ok {
		return usi, err
	}

	return usi, nil
}

// isValid checks the fullname and SHAs
func (usi *UpdateServiceItem) isValid() (bool, error) {
	if usi.FullName == "" || len(usi.SHAS) == 0 {
		return false, errors.New("Fullname/SHA256 fields should not be empty")
	}

	return true, nil
}

// Equal compares fullname to see if these two items are the same
func (usi *UpdateServiceItem) Equal(item UpdateServiceItem) bool {
	return usi.FullName == item.FullName
}

// IsExpired tells if an application is expired
func (usi *UpdateServiceItem) IsExpired() bool {
	return usi.Expired.Before(time.Now())
}

// UpdateServiceMeta represents the meta info of a repository
type UpdateServiceMeta struct {
	Proto      string
	Namespace  string
	Repository string
	Items      []UpdateServiceItem
	Updated    time.Time
}

// NewUpdateServiceMeta creates/loads a UpdateServiceMeta by 'proto', 'namespace' and 'repository'
func NewUpdateServiceMeta(p, n, r string) (UpdateServiceMeta, error) {
	if p == "" || n == "" || r == "" {
		return UpdateServiceMeta{}, errors.New("Fail to create a meta with empty Proto/Namespace/Repository")
	}

	var usm UpdateServiceMeta
	topDir := filepath.Join(setting.DockyardPath, p, defaultMetaDirName, n, r)
	metaFile := filepath.Join(topDir, defaultMetaFileName)
	if utils.IsFileExist(metaFile) {
		data, err := ioutil.ReadFile(metaFile)
		if err != nil {
			return UpdateServiceMeta{}, err
		}

		if err := json.Unmarshal(data, &usm); err != nil {
			return UpdateServiceMeta{}, err
		}

		return usm, nil
	}
	usm.Proto = p
	usm.Namespace = n
	usm.Repository = r

	// provides empty meta
	usm.save()
	return usm, nil
}

// ReadMeta provides meta bytes for
func (usm *UpdateServiceMeta) ReadMeta() ([]byte, error) {
	metaFile := filepath.Join(usm.getTopDir(), defaultMetaFileName)
	return ioutil.ReadFile(metaFile)
}

func (usm *UpdateServiceMeta) ReadMetaSign() ([]byte, error) {
	signFile := filepath.Join(usm.getTopDir(), defaultMetaSignFileName)
	return ioutil.ReadFile(signFile)
}

// Get gets an UpdateServiceItem by 'fullname'
func (usm *UpdateServiceMeta) Get(fullname string) (UpdateServiceItem, error) {
	if usm.Proto == "" || usm.Namespace == "" || usm.Repository == "" {
		return UpdateServiceItem{}, errors.New("Fail to get a meta with empty Proto/Namespace/Repository")
	}

	if fullname == "" {
		return UpdateServiceItem{}, errors.New("'FullName' should not be empty")
	}

	for _, item := range usm.Items {
		if item.FullName == fullname {
			return item, nil
		}
	}

	return UpdateServiceItem{}, fmt.Errorf("Cannot find the meta item: %s", fullname)
}

// Put adds an UpdateServiceItem to meta data, save both meta file and sign file
func (usm *UpdateServiceMeta) Put(usi UpdateServiceItem) error {
	if usm.Proto == "" || usm.Namespace == "" || usm.Repository == "" {
		return errors.New("Fail to put a meta with empty Proto/Namespace/Repository")
	}

	exist := false
	for i := range usm.Items {
		if usm.Items[i].Equal(usi) {
			usm.Items[i] = usi
			exist = true
		}
	}
	if !exist {
		usm.Items = append(usm.Items, usi)
	}

	if err := usm.save(); err != nil {
		return err
	}

	return nil
}

// Delete removes an UpdateServiceItem from meta data, save both meta file and sign file after that
func (usm *UpdateServiceMeta) Delete(fullname string) error {
	if usm.Proto == "" || usm.Namespace == "" || usm.Repository == "" {
		return errors.New("Fail to remove a meta with empty Proto/Namespace/Repository")
	}

	exist := false
	for i := range usm.Items {
		if usm.Items[i].FullName == fullname {
			usm.Items = append(usm.Items[:i], usm.Items[i+1:]...)
			exist = true
			break
		}
	}

	if !exist {
		return errors.New("Cannot find the meta item")
	}

	if err := usm.save(); err != nil {
		return err
	}

	return nil
}

// getTopDir provides an inner way to compose a top directory of a meta
func (usm *UpdateServiceMeta) getTopDir() string {
	return filepath.Join(setting.DockyardPath, usm.Proto, defaultMetaDirName, usm.Namespace, usm.Repository)
}

// save saves meta data to local file
func (usm *UpdateServiceMeta) save() error {
	topDir := usm.getTopDir()
	usm.Updated = time.Now()
	usmContent, _ := json.Marshal(usm)

	if !utils.IsDirExist(topDir) {
		err := os.MkdirAll(topDir, 0755)
		if err != nil {
			return err
		}
	}

	metaMutex.Lock()
	defer metaMutex.Unlock()
	metaFile := filepath.Join(topDir, defaultMetaFileName)
	err := ioutil.WriteFile(metaFile, usmContent, 0644)
	if err != nil {
		return err
	}

	err = KeyManagerEnabled()
	if err == nil {
		// write sign file, don't popup error even fail to saveSign
		usm.saveSign(usmContent)
	}

	return nil
}

// saveSign signs the meta data and save the signed data to local file
func (usm *UpdateServiceMeta) saveSign(content []byte) error {
	mode, err := NewKeyManagerMode(setting.KeyManagerMode)
	if err != nil {
		return err
	}
	manager, err := NewKeyManager(mode, setting.KeyManagerURI)
	if err != nil {
		return err
	}
	signContent, err := manager.Sign(usm.Proto, usm.Namespace, content)
	if err != nil {
		return err
	}
	signFile := filepath.Join(usm.getTopDir(), defaultMetaSignFileName)
	if err := ioutil.WriteFile(signFile, signContent, 0644); err != nil {
		return err
	}

	return nil
}
