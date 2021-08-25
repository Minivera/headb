package keys

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyLength    = 12
	encryptCost     = 10
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

// EncryptToPaseto will merge the given API key string with a database integer ID using
// a paseto web token and return the encrypted key.
func EncryptToPaseto(key string, keyID int64, encryptionSecret string) (string, error) {
	symmetricKey := []byte(encryptionSecret)
	now := time.Now()

	jsonToken := paseto.JSONToken{
		Issuer:     "headb",
		// Expires in 99 years
		Expiration: now.Add(99 * time.Hour * 365),
	}

	jsonToken.Set("key_value", key)
	jsonToken.Set("key_id", strconv.FormatInt(keyID, 10))

	token, err := paseto.NewV2().Encrypt(symmetricKey, jsonToken, "")
	if err != nil {
		return "", err
	}

	return token, nil
}

// ExtractIDAndValue extracts the database ID and API key value from
// an encoded API key string through paseto.
func ExtractIDAndValue(mergedKey string, encryptionSecret string) (string, int64, error) {
	var newJsonToken paseto.JSONToken
	var newFooter string
	err := paseto.NewV2().Decrypt(mergedKey, []byte(encryptionSecret), &newJsonToken, &newFooter)
	if err != nil {
		return "", 0, err
	}

	keyID, err := strconv.ParseInt(newJsonToken.Get("key_id"), 10, 64)
	if err != nil {
		return "", 0, err
	}

	return newJsonToken.Get("key_value"), keyID, nil
}

// ValidateKey validates that the plain API key given is the plain version of
// the hashed API key. This can be used to compare user facing keys with keys
// we save in the database.
func ValidateKey(plainKey, hashedKey string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(plainKey))
}
