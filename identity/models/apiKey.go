package models

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

type ApiKey struct {
	ID         uint64
	Value      string
	UserID     uint64
	LastUsedAt time.Time
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

func NewApiKey(value string, userID uint64) *ApiKey {
	return &ApiKey{
		Value:  value,
		UserID: userID,
	}
}

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
	err := sqldb.QueryRow(ctx, keyQuery, ID).Scan(&key.ID, &key.Value, &key.UserID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.Errorf("Could not query api_key for ID, %v", err)
		return nil, err
	}

	return &key, nil
}

func (key *ApiKey) Save(ctx context.Context) error {
	keyQuery := `
		INSERT INTO "api_keys" (value, user_id)
		VALUES ($1, $2)
		RETURNING id, last_used_at, updated_at, created_at;
	`

	err := sqldb.QueryRow(ctx, keyQuery, key.Value, key.UserID).Scan(&key.ID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.Errorf("Could not insert key (API key update is not supported), %v", err)
		return err
	}

	return nil
}

func (key *ApiKey) Delete(ctx context.Context) error {
	keyQuery := `
		DELETE FROM "api_keys"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, keyQuery, key.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.Errorf("Could not delete key, %v", err)
		return err
	}

	return nil
}
