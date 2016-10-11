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

package utils

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// IsDirExist checks if a dir exists
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}

	return fi.IsDir()
}

// IsFileExist checks if a file exists
func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// MD5 creates md5 string for an input key
func MD5(key string) string {
	md5String := fmt.Sprintf("dockyard %s is a container %d hub", key, time.Now().Unix())
	h := md5.New()
	h.Write([]byte(md5String))

	return hex.EncodeToString(h.Sum(nil))
}

// SHA512 creates sha512 string for an input data
func SHA512(body []byte) (string, error) {
	sha512h := sha512.New()
	_, err := io.Copy(sha512h, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha512h.Sum(nil)), nil
}

// SHA512FromFile creates sha512 string for a file
func SHA512FromStream(reader io.Reader) (string, error, int64) {
	sha512h := sha512.New()
	n, err := io.Copy(sha512h, reader)
	if err != nil {
		return "", err, 0
	}

	hash512 := fmt.Sprintf("%x", sha512h.Sum(nil))
	return hash512, nil, n
}

// Compare returns 0 if a, b are equal, -1 if a < b, other wise returns 1
func Compare(a, b string) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return +1
}

// GenerateRSAKeyPair generate a private key and a public key
func GenerateRSAKeyPair(bits int) ([]byte, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)

	if err != nil {
		return nil, nil, err
	}

	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes}

	return pem.EncodeToMemory(privBlock), pem.EncodeToMemory(pubBlock), nil
}

// RSAEncrypt encrypts a content by a public key
func RSAEncrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, contentBytes)
}

// RSADecrypt decrypts content by a private key
func RSADecrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privKey, contentBytes)
}

// SHA256Sign signs a content by a private key
func SHA256Sign(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256(contentBytes)
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

// SHA256Verify verifies if a content is valid by a signed data an a public key
func SHA256Verify(keyBytes []byte, contentBytes []byte, signBytes []byte) error {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return err
	}

	signStr := hex.EncodeToString(signBytes)
	newSignBytes, _ := hex.DecodeString(signStr)
	hashed := sha256.Sum256(contentBytes)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], newSignBytes)
}

func getPrivKey(privBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("Fail to decode private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func getPubKey(pubBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return nil, errors.New("Fail to decode public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Fail get public key from public interface")
	}

	return pubKey, nil
}
