package models

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

// ApiKey is a struct that mirrors the API key table
// from the database.
type ApiKey struct {
	ID         uint64
	Value      string
	UserID     uint64
	LastUsedAt time.Time
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

// NewApiKey generates a new API key structure from an encrypted key
// value and a user ID. Never give plain API key values to this function.
func NewApiKey(value string, userID uint64) *ApiKey {
	return &ApiKey{
		Value:  value,
		UserID: userID,
	}
}

// GetApiKey gets an API key from a database ID, returning it
// if successfully fetched.
func GetApiKey(ctx context.Context, ID uint64) (*ApiKey, error) {
	keyQuery := `
		SELECT
			id,
			value,
			user_id,
			last_used_at,
			updated_at,
			created_at
		FROM
			"api_keys"
		WHERE
			api_keys.id = $1
		LIMIT 1;
	`

	key := ApiKey{}
	err := sqldb.
		QueryRow(ctx, keyQuery, ID).
		Scan(&key.ID, &key.Value, &key.UserID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not query api_key for ID")
		return nil, err
	}

	return &key, nil
}

// Save saves the data of the key it used on. This method only saves
// the value and user ID from the struct and updates the timestamps. This method
// saves the value as-is, always make sure to encrypt the API key before
// saving. Save will update only the LastUserAt field when the key already exists.
func (key *ApiKey) Save(ctx context.Context) error {
	keyQuery := `
		INSERT INTO "api_keys" (value, user_id, last_used_at)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT value_user_id_unique DO UPDATE SET last_used_at = $3
		RETURNING id, last_used_at, updated_at, created_at;
	`

	err := sqldb.
		QueryRow(ctx, keyQuery, key.Value, key.UserID, key.LastUsedAt).
		Scan(&key.ID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not insert key (API key update is not supported)")
		return err
	}

	return nil
}

// Delete deletes the API key is it called on.
func (key *ApiKey) Delete(ctx context.Context) error {
	keyQuery := `
		DELETE FROM "api_keys"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, keyQuery, key.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Errorf("Could not delete key")
		return err
	}

	return nil
}
