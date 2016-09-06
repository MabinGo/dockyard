/*
Copyright 2016 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Mode 2.0 (the "License");
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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
)

func TestNewUpdateServiceItem(t *testing.T) {
	cases := []struct {
		fn       string
		shas     []string
		expected bool
	}{
		{"fn", []string{"sha1"}, true},
		{"", []string{"sha1"}, false},
		{"fn", []string{}, false},
	}
	for _, c := range cases {
		_, err := NewUpdateServiceItem(c.fn, c.shas)
		assert.Equal(t, c.expected, err == nil, "Fail to create new item")
	}
}

func TestUpdateServiceItemEqual(t *testing.T) {
	cases := []struct {
		fn       string
		shas     []string
		expected bool
	}{
		{"fn", []string{"sha1"}, true},
		{"fn1", []string{"sha3"}, false},
	}

	testItem, _ := NewUpdateServiceItem("fn", []string{"sha0"})
	for _, c := range cases {
		cItem, _ := NewUpdateServiceItem(c.fn, c.shas)
		assert.Equal(t, c.expected, testItem.Equal(cItem))
	}
}

func TestUpdateServiceItemIsExpired(t *testing.T) {
	testItem, _ := NewUpdateServiceItem("fn", []string{"sha0"})

	testNewExpired := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	testItem.Expired = testNewExpired
	assert.Equal(t, true, testItem.IsExpired(), "Fail to get the expired status")
	testItem.Expired = time.Now().Add(time.Hour * 1)
	assert.Equal(t, false, testItem.IsExpired(), "Fail to get the expired status")

	testNewCreated := testItem.Created.Add(time.Hour * 2)
	testItem.Created = testNewCreated

	assert.Equal(t, 1, len(testItem.SHAS), "Fail to get correct shas count")
	assert.Equal(t, "sha0", testItem.SHAS[0], "Fail to get correct shas")
}

func TestUpdateServiceMeta(t *testing.T) {
	tmpHome, err := ioutil.TempDir("", "usm-test-")
	assert.Nil(t, err, "Fail to create a temp dir")
	defer os.RemoveAll(tmpHome)

	// setup a temp DockyardPath and KeyManagerURI
	setting.DockyardPath = filepath.Join(tmpHome, "storage")
	setting.KeyManagerURI = filepath.Join(tmpHome, "km")

	for _, mode := range []string{"single", "share"} {
		setting.KeyManagerMode = mode

		_, err = NewUpdateServiceMeta("", "", "")
		assert.NotNil(t, err, "Should not create meta by invalid item")

		// add an 'fn/sha0' item
		testMeta, err := NewUpdateServiceMeta("p", "n", "r")
		assert.Nil(t, err, "Fail to init update service meta")
		testItem, _ := NewUpdateServiceItem("fn", []string{"sha0"})
		err = testMeta.Put(testItem)
		assert.Nil(t, err, "Fail to add a test item")

		// query an 'fn' item and compare it
		newMeta, _ := NewUpdateServiceMeta("p", "n", "r")
		_, err = newMeta.Get("invalidfn")
		assert.NotNil(t, err, "Should not load item with invalid fullname")
		retItem, err := newMeta.Get("fn")
		assert.Nil(t, err, "Fail to load exist item")
		assert.Equal(t, testItem.FullName, retItem.FullName, "Fail to load the correct fullname")
		assert.Equal(t, len(testItem.SHAS), len(retItem.SHAS), "Fail to load the correct SHAS count")
		assert.Equal(t, testItem.SHAS[0], retItem.SHAS[0], "Fail to load the correct SHAS value")

		// update 'fn' item with 'sha1'
		updatedItem, _ := NewUpdateServiceItem("fn", []string{"sha0-updated"})
		err = newMeta.Put(updatedItem)
		assert.Nil(t, err, "Fail to update a test item")
		newUpdatedMeta, _ := NewUpdateServiceMeta("p", "n", "r")
		retUpdatedItem, err := newUpdatedMeta.Get("fn")
		assert.Nil(t, err, "Fail to load exist item")
		assert.Equal(t, len(updatedItem.SHAS), len(retUpdatedItem.SHAS), "Fail to load the correct SHAS count")
		assert.Equal(t, updatedItem.SHAS[0], retUpdatedItem.SHAS[0], "Fail to load the correct SHAS value")

		// read meta file
		_, err = newMeta.ReadMeta()
		assert.Nil(t, err, "Fail to read meta file")

		// read meta sign file
		_, err = newMeta.ReadMetaSign()
		assert.Nil(t, err, "Fail to read meta sign")

		err = newMeta.Delete("fn")
		assert.Nil(t, err, "Fail to delete meta item")
		err = newMeta.Delete("fn")
		assert.NotNil(t, err, "Should return error in deleting non exist item")
	}
}

func TestUpdateService(t *testing.T) {
	tmpHome, err := ioutil.TempDir("", "us-test-")
	assert.Nil(t, err, "Fail to create a temp dir")
	defer os.RemoveAll(tmpHome)

	// setup a temp DockyardPath and KeyManagerURI
	setting.DockyardPath = filepath.Join(tmpHome, "storage")
	setting.KeyManagerURI = filepath.Join(tmpHome, "km")

	for _, mode := range []string{"single", "share"} {
		setting.KeyManagerMode = mode

		var upService UpdateService
		err := upService.Put("p", "n", "r", "fn", []string{"sha0"})
		assert.Nil(t, err, "Fail to put in update service")
		metaBytes, err := upService.ReadMeta("p", "n", "r")
		assert.Nil(t, err, "Fail to read meta data")
		signBytes, err := upService.ReadMetaSign("p", "n", "r")
		assert.Nil(t, err, "Fail to read meta sign")
		pubBytes, err := upService.ReadPublicKey("p", "n")
		assert.Nil(t, err, "Fail to read public key")

		assert.Nil(t, utils.SHA256Verify(pubBytes, metaBytes, signBytes), "Fail to verify update service metadata")

		err = upService.Delete("p", "n", "r", "fn")
		assert.Nil(t, err, "Fail to delete item from update service")
		err = upService.Delete("p", "n", "r", "fn")
		assert.NotNil(t, err, "Should return error in deleting non exist item")
	}

}

func TestUpdateServiceKMDisabled(t *testing.T) {
	tmpHome, err := ioutil.TempDir("", "us-test-")
	assert.Nil(t, err, "Fail to create a temp dir")
	defer os.RemoveAll(tmpHome)

	// setup a temp DockyardPath and KeyManagerURI
	setting.DockyardPath = filepath.Join(tmpHome, "storage")
	setting.KeyManagerURI = ""

	for _, mode := range []string{"single", "share"} {
		setting.KeyManagerMode = mode

		var upService UpdateService
		err := upService.Put("p", "n", "r", "fn", []string{"sha0"})
		assert.Nil(t, err, "Fail to put in update service")
		_, err = upService.ReadMeta("p", "n", "r")
		assert.Nil(t, err, "Fail to read meta data")
		_, err = upService.ReadMetaSign("p", "n", "r")
		assert.NotNil(t, err, "Should not read meta sign without key manager enabled")
		_, err = upService.ReadPublicKey("p", "n")
		assert.NotNil(t, err, "Should to read public key without key manager enabled")
	}

}
