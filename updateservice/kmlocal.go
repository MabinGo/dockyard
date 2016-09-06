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
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/containerops/dockyard/utils"
)

const (
	// LocalPrefix represents this implementation name
	LocalPrefix       = "local"
	defaultPublicKey  = "pub_key.pem"
	defaultPrivateKey = "priv_key.pem"
	defaultBitsSize   = 2048
)

var (
	keyMutex sync.Mutex
)

// KeyManagerLocal is the local implementation of a key manager
type KeyManagerLocal struct {
	Mode KeyManagerMode
	Path string
}

// init regists to the global KeyManager
func init() {
	RegisterKeyManager(LocalPrefix, &KeyManagerLocal{})
}

// Supported checks if an uri is simply a local path like "/data"
func (kml *KeyManagerLocal) Supported(mode KeyManagerMode, uri string) bool {
	if uri == "" {
		return false
	}

	if mode != Single && mode != Share {
		return false
	}

	if u, err := url.Parse(uri); err != nil {
		return false
	} else if u.Scheme == "" {
		return true
	}

	return false
}

// New returns a keymanager by an uri
func (kml *KeyManagerLocal) New(mode KeyManagerMode, uri string) (KeyManager, error) {
	if !kml.Supported(mode, uri) {
		return nil, errors.New("Fail to create a key manager local interface, the uri is not supported")
	}

	kml.Mode = mode
	kml.Path = uri
	return kml, nil
}

// getKeyDir provides an inner way to compose a key dir of the local key manager
func (kml *KeyManagerLocal) getKeyDir(proto, namespace string) (string, error) {
	var keyDir string
	switch kml.Mode {
	case Share:
		keyDir = filepath.Join(kml.Path)
	case Single:
		keyDir = filepath.Join(kml.Path, proto, namespace)
	default:
		return "", fmt.Errorf("Key manager mode <%s> is not supported in local implement.", string(kml.Mode))
	}

	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return "", err
		}
	}

	return keyDir, nil
}

// ReadPublicKey gets the public key data of a namespace
func (kml *KeyManagerLocal) ReadPublicKey(proto string, namespace string) ([]byte, error) {
	keyDir, err := kml.getKeyDir(proto, namespace)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(keyDir, defaultPublicKey))
}

// Decrypt decrypts a data of a namespace
func (kml *KeyManagerLocal) Decrypt(proto string, namespace string, data []byte) ([]byte, error) {
	keyDir, err := kml.getKeyDir(proto, namespace)
	if err != nil {
		return nil, err
	}

	privBytes, err := ioutil.ReadFile(filepath.Join(keyDir, defaultPrivateKey))
	if err != nil {
		return nil, err
	}
	return utils.RSADecrypt(privBytes, data)
}

// Sign signs a data of a namespace
func (kml *KeyManagerLocal) Sign(proto string, namespace string, data []byte) ([]byte, error) {
	keyDir, err := kml.getKeyDir(proto, namespace)
	if err != nil {
		return nil, err
	}

	privBytes, err := ioutil.ReadFile(filepath.Join(keyDir, defaultPrivateKey))
	if err != nil {
		return nil, err
	}
	return utils.SHA256Sign(privBytes, data)
}

// isKeyExist provides an inner way to check if key pairs are both exist
func isKeyExist(keyDir string) bool {
	if !utils.IsFileExist(filepath.Join(keyDir, defaultPrivateKey)) {
		return false
	}

	if !utils.IsFileExist(filepath.Join(keyDir, defaultPublicKey)) {
		return false
	}

	return true
}

// generateKey provides an inner way to generate a RSA key pairs
func generateKey(keyDir string) error {
	privBytes, pubBytes, err := utils.GenerateRSAKeyPair(defaultBitsSize)
	if err != nil {
		return err
	}

	if !utils.IsDirExist(keyDir) {
		err := os.MkdirAll(keyDir, 0755)
		if err != nil {
			return err
		}
	}

	keyMutex.Lock()
	defer keyMutex.Unlock()
	if err := ioutil.WriteFile(filepath.Join(keyDir, defaultPrivateKey), privBytes, 0600); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(keyDir, defaultPublicKey), pubBytes, 0644); err != nil {
		return err
	}

	return nil
}
