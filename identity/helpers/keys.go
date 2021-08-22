package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyLength    = 24
	encryptCost     = 10
	apiKeySeparator = "."
)

// GenerateApiKey will generate a new unencrypted, unencoded API key. This key it to be
// used only internally.
func GenerateApiKey() (string, error) {
	key := make([]byte, apiKeyLength)

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(key), nil
}

// GenerateSecureApiKey generates a secure version of the given API key string
// this key can be saved in the database as it is securely encrypted.
func GenerateSecureApiKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), encryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// MergeWithKeyID will merge the given API key string with a database integer ID using
// the API key separator.
func MergeWithKeyID(key string, keyID uint64) string {
	keyIDbase64 := base64.URLEncoding.EncodeToString([]byte(strconv.FormatUint(keyID, 10)))
	return fmt.Sprintf("%s%s%s", keyIDbase64, apiKeySeparator, key)
}

// ExtractIDAndValue extracts the database ID and API key value from
// an encoded API key string.
func ExtractIDAndValue(mergedKey string) (string, uint64, error) {
	parts := strings.Split(mergedKey, ".")
	if len(parts) < 2 || len(parts) > 2 {
		return "", 0, fmt.Errorf("received API key was not a two part key")
	}

	keyIDbase64 := parts[0]
	keyIDString, err := base64.URLEncoding.DecodeString(keyIDbase64)
	if err != nil {
		return "", 0, err
	}

	keyID, err := strconv.ParseUint(string(keyIDString), 10, 64)
	if err != nil {
		return "", 0, err
	}

	return parts[1], keyID, nil
}

// ValidateKey validates that the plain API key given is the plain version of
// the hashed API key. This can be used to compare user facing keys with keys
// we save in the database.
func ValidateKey(plainKey, hashedKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(plainKey))
}
