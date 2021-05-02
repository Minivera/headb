package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	ApiKeyLength    = 24
	EncryptCost     = 14
	ApiKeySeparator = "."
)

func GenerateApiKey() (string, error) {
	key := make([]byte, ApiKeyLength)

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(key), nil
}

func GenerateSecureApiKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), EncryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func MergeWithKeyID(key string, keyID uint64) string {
	keyIDbase64 := base64.URLEncoding.EncodeToString([]byte(strconv.FormatUint(keyID, 10)))
	return fmt.Sprintf("%s%s%s", key, ApiKeySeparator, keyIDbase64)
}

func ExtractIDAndValue(mergedKey string) (string, uint64, error) {
	parts := strings.Split(mergedKey, ".")
	if len(parts) < 2 || len(parts) > 2 {
		return "", 0, fmt.Errorf("received API key was not a two part key")
	}

	keyIDbase64 := parts[1]
	keyIDString, err := base64.URLEncoding.DecodeString(keyIDbase64)
	if err != nil {
		return "", 0, err
	}

	keyID, err := strconv.ParseUint(string(keyIDString), 10, 64)
	if err != nil {
		return "", 0, err
	}

	return parts[0], keyID, nil
}

func ValidateKey(plainKey, hashedKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(plainKey))
}
