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
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestKMLData(mode KeyManagerMode, t *testing.T) (KeyManager, string) {
	var local KeyManagerLocal
	_, path, _, _ := runtime.Caller(0)
	testFilePath := filepath.Join(filepath.Dir(path), "../tests/unittestdata")

	l, err := local.New(mode, filepath.Join(testFilePath, "km", string(mode)))
	assert.Nil(t, err, "Fail to setup a local test key manager")

	return l, testFilePath
}

func TestKMLSupportedBasic(t *testing.T) {
	var local KeyManagerLocal
	cases := []struct {
		mode     KeyManagerMode
		url      string
		expected bool
	}{
		{Share, "/tmp/", true},
		{Single, "/tmp/", true},
		{None, "/tmp/", false},
		{Share, "", false},
		{Share, "invalid://tmp", false},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, local.Supported(c.mode, c.url), "Fail to get supported status")
	}
}

func TestKMLNew(t *testing.T) {
	var local KeyManagerLocal
	cases := []struct {
		mode     KeyManagerMode
		url      string
		expected bool
	}{
		{Share, "/tmp/", true},
		{Share, "", false},
	}

	for _, c := range cases {
		_, err := local.New(c.mode, c.url)
		assert.Equal(t, c.expected, err == nil, "Fail to create a new local key manager")
	}
}

func TestKMLReadPublicKey(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "us-test-")
	defer os.RemoveAll(tmpPath)
	assert.Nil(t, err, "Fail to create temp dir")

	cases := []struct {
		mode     KeyManagerMode
		url      string
		p        string
		n        string
		expected bool
	}{
		{Share, tmpPath, "p", "n", true},
		{Single, tmpPath, "p", "n", true},
		{None, tmpPath, "p", "n", false},
	}
	for _, c := range cases {
		var local KeyManagerLocal
		local.Mode = c.mode
		local.Path = c.url

		_, err = local.ReadPublicKey(c.p, c.n)
		assert.Equal(t, c.expected, err == nil, "Fail to read public key")
	}
}

func TestKMLSign(t *testing.T) {
	proto := "app"
	namespace := "containerops"
	for _, m := range []KeyManagerMode{Share, Single} {
		l, testFilePath := loadTestKMLData(m, t)
		testFile := filepath.Join(testFilePath, "hello.txt")
		testBytes, _ := ioutil.ReadFile(testFile)
		signFile := filepath.Join(testFilePath, "hello.sig")
		signBytes, _ := ioutil.ReadFile(signFile)

		data, err := l.Sign(proto, namespace, testBytes)
		assert.Nil(t, err, "Fail to sign")
		assert.Equal(t, data, signBytes, "Fail to sign correctly")
	}
}

func TestKMLDecrypt(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	testFilePath := filepath.Join(filepath.Dir(path), "../tests/unittestdata")

	cases := []struct {
		mode         KeyManagerMode
		url          string
		p            string
		n            string
		testFile     string
		encryptFile  string
		errExpected  bool
		dataExpected bool
	}{
		{Share, testFilePath, "app", "containerops", "hello.txt", "hello.encrypt", true, true},
		{Single, testFilePath, "app", "containerops", "hello.txt", "hello.encrypt", true, true},
		{None, testFilePath, "app", "containerops", "hello.txt", "hello.encrypt", false, false},
		{Share, filepath.Join(testFilePath, "hello.txt"), "app", "containerops", "hello.txt", "hello.encrypt", false, false},
		{Single, testFilePath, "app", "containerops", "hello.encrypt", "hello.encrypt", true, false},
	}

	for _, c := range cases {
		var local KeyManagerLocal
		local.Mode = c.mode
		local.Path = filepath.Join(c.url, "km", string(c.mode))
		testFile := filepath.Join(testFilePath, c.testFile)
		testBytes, _ := ioutil.ReadFile(testFile)
		testEncryptedFile := filepath.Join(testFilePath, c.encryptFile)
		testEncryptedBytes, _ := ioutil.ReadFile(testEncryptedFile)

		testDecryptedBytes, err := local.Decrypt(c.p, c.n, testEncryptedBytes)
		assert.Equal(t, c.errExpected, err == nil, "Fail to decrypt")
		if c.dataExpected {
			assert.Equal(t, testDecryptedBytes, testBytes, "Fail to decrypt correctly")
		} else {
			assert.NotEqual(t, testDecryptedBytes, testBytes, "Fail to decrypt correctly")
		}
	}
}

func testKMLgenerateKey(t *testing.T) {
	tmpPath, _ := ioutil.TempDir("", "us-test-")
	defer os.RemoveAll(tmpPath)

	assert.Nil(t, generateKey(tmpPath), "Fail to generate key")
}

func TestKMLisKeyExist(t *testing.T) {
	tmpPath, _ := ioutil.TempDir("", "us-test-")
	defer os.RemoveAll(tmpPath)

	generateKey(tmpPath)
	assert.Equal(t, true, isKeyExist(tmpPath), "Fail to check valid path")
	assert.Equal(t, false, isKeyExist(filepath.Join(tmpPath, "invalid")), "Fail to check empty path")
}
