package github

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// EncryptAccessToken will encrypt a GitHub access token through a cypher so
// it can safely be saved in the database.
func EncryptAccessToken(data, passphrase string) (string, error) {
	cipherBlock, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(data), nil)), nil
}

// DecryptAccessToken decrypts an encrypted access token using the same
// passphrase as the encrypted version.
func DecryptAccessToken(data, passphrase string) (string, error) {
	encryptData, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	cipherBlock, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(encryptData) < nonceSize {
		return "", err
	}

	nonce, cipherText := encryptData[:nonceSize], encryptData[nonceSize:]
	plainData, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainData), nil
}
