package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// AESKey is 32 bytes aes key.
var AESKey = []byte{0x9E, 0xDC, 0x22, 0xAB, 0x4B, 0xDD, 0x78, 0xF5,
	0xA1, 0x45, 0xFE, 0x61, 0x1D, 0x2C, 0xB3, 0xD2, 0xF4, 0x15, 0x6B,
	0x60, 0x9F, 0xE5, 0x45, 0xE1, 0xA3, 0x80, 0x79, 0xD8, 0xED, 0x98,
	0x1C, 0xD2}

// Encrypt will encrypt the message with cipher feedback mode with the given key.
func Encrypt(message, AESKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(message))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], message)
	return ciphertext, nil
}

// Decrypt will decrypt the ciphertext via the given key.
func Decrypt(ciphertext, AESKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}
