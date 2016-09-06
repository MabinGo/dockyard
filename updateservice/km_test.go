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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/containerops/dockyard/setting"
)

type KeyManagerFake struct{}

func (f *KeyManagerFake) New(mode KeyManagerMode, url string) (KeyManager, error) {
	if url != "fake://empty" {
		return f, errors.New("Test error")
	}
	return f, nil
}
func (f *KeyManagerFake) Supported(mode KeyManagerMode, url string) bool {
	if mode != Share {
		return false
	}
	if !strings.HasPrefix(url, "fake") {
		return false
	}
	return true
}
func (f *KeyManagerFake) ReadPublicKey(proto string, namespace string) ([]byte, error) {
	return nil, nil
}
func (f *KeyManagerFake) Sign(proto string, namespace string, data []byte) ([]byte, error) {
	return nil, nil
}

func (f *KeyManagerFake) Decrypt(proto string, namespace string, data []byte) ([]byte, error) {
	return nil, nil
}

func TestNewKeyManager(t *testing.T) {
	RegisterKeyManager("testfakename", &KeyManagerFake{})
	cases := []struct {
		mode     KeyManagerMode
		url      string
		expected bool
	}{
		{Share, "fake://empty", true},
		{Single, "fake://empty", false},
		{Share, "fake://error", false},
		{Share, "", false},
		{Share, "error://", false},
	}
	for _, c := range cases {
		_, err := NewKeyManager(c.mode, c.url)
		assert.Equal(t, c.expected, err == nil, "Error in creating key manager")
	}
}

func TestRegisterKeyManager(t *testing.T) {
	cases := []struct {
		name     string
		f        KeyManager
		expected bool
	}{
		{"", &KeyManagerLocal{}, false},
		{"testkmname", &KeyManagerLocal{}, true},
		{"testkmname", &KeyManagerLocal{}, false},
		{"testkmname2", nil, false},
	}

	for _, c := range cases {
		err := RegisterKeyManager(c.name, c.f)
		t.Log(err, c)
		assert.Equal(t, c.expected, err == nil, "Fail to register key manager")
	}
}

func TestKMMode(t *testing.T) {
	cases := []struct {
		str      string
		expected bool
	}{
		{"share", true},
		{"single", true},
		{"db", false},
		{"", false},
	}

	for _, c := range cases {
		_, err := NewKeyManagerMode(c.str)
		assert.Equal(t, c.expected, err == nil, "Fail to create key manager mode by string")
	}
}

func TestKMEnable(t *testing.T) {
	cases := []struct {
		mode     string
		url      string
		expected bool
	}{
		{"share", "/tmp", true},
		{"bad", "/tmp", false},
		{"share", "", false},
		{"share", "unknown://data", false},
	}

	for _, c := range cases {
		setting.KeyManagerMode = c.mode
		setting.KeyManagerURI = c.url
		err := KeyManagerEnabled()
		assert.Equal(t, c.expected, err == nil, "Fail to get keymanager enabled status")
	}
}
