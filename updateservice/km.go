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
	"fmt"
	"sync"

	"github.com/containerops/dockyard/setting"
)

// KeyManagerMode presents a supported key manager mode
type KeyManagerMode string

const (
	// Share mode: all the namespace share a same key pair
	Share KeyManagerMode = "share"
	// Single mode: each namespace maintain its own key pair
	Single KeyManagerMode = "single"
	// None mode: invalid/non supported
	None KeyManagerMode = "none"
)

// NewKeyManagerMode returns a mode from a string
func NewKeyManagerMode(mode string) (KeyManagerMode, error) {
	switch mode {
	case "share":
		return Share, nil
	case "single":
		return Single, nil
	default:
		return None, fmt.Errorf("The key manager mode <%s> is invalid.", mode)
	}
}

// KeyManager should be seperate from dockyard
// Now only assume that keys are existed in the backend key manager.
// It is up to each implementation to decide whether provides a way
//  to generate key pair automatically.
// NOTE: it is dangrous to privide a ReadPrivateKey interface
type KeyManager interface {
	// `url` is the database address or local directory (/tmp/cache)
	New(mode KeyManagerMode, url string) (KeyManager, error)
	Supported(mode KeyManagerMode, url string) bool
	// proto: 'app/v1' for example
	ReadPublicKey(proto string, namespace string) ([]byte, error)
	// proto: 'app/v1' for example
	Sign(proto string, namespace string, data []byte) ([]byte, error)
	// proto: 'app/v1' for example
	Decrypt(proto string, namespace string, data []byte) ([]byte, error)
}

var (
	kmsLock sync.Mutex
	kms     = make(map[string]KeyManager)

	// ErrorsKMNotSupported occurs when the km type is not supported
	ErrorsKMNotSupported = errors.New("key manager type is not supported")
	// ErrorsKMNotEnabled occurs when the km type is not enabled
	ErrorsKMNotEnabled = errors.New("key manager type is not enabled")
)

// RegisterKeyManager provides a way to dynamically register an implementation of a
// key manager type.
//
// If RegisterKeyManager is called twice with the same name if 'keymanager type' is nil,
// or if the name is blank, it panics.
func RegisterKeyManager(name string, f KeyManager) error {
	if name == "" {
		return errors.New("Could not register a KeyManager with an empty name")
	}
	if f == nil {
		return errors.New("Could not register a nil KeyManager")
	}

	kmsLock.Lock()
	defer kmsLock.Unlock()

	if _, alreadyExists := kms[name]; alreadyExists {
		return fmt.Errorf("KeyManager type '%s' is already registered", name)
	}
	kms[name] = f

	return nil
}

func KeyManagerEnabled() error {
	if setting.KeyManagerURI == "" {
		return ErrorsKMNotEnabled
	}

	mode, err := NewKeyManagerMode(setting.KeyManagerMode)
	if err != nil {
		return err
	}
	_, err = NewKeyManager(mode, setting.KeyManagerURI)
	if err != nil {
		return err
	}

	return nil
}

// NewKeyManager create a key manager by a url
func NewKeyManager(mode KeyManagerMode, url string) (KeyManager, error) {
	for _, f := range kms {
		if f.Supported(mode, url) {
			return f.New(mode, url)
		}
	}

	return nil, ErrorsKMNotSupported
}
