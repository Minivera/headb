package models

import (
	"context"

	"encore.dev/storage/sqldb"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
)

// NewApiKey generates a new API key structure from an encrypted key
// value and a user ID. Never give plain API key values to this function.
func NewApiKey(value string, userID int64) *model.APIKeys {
	return &model.APIKeys{
		Value:  value,
		UserID: userID,
	}
}

// GetApiKey gets an API key from a database ID, returning it
// if successfully fetched.
func GetApiKey(ctx context.Context, ID int64) (*model.APIKeys, error) {
	query, args := postgres.SELECT(
		table.APIKeys.ID,
		table.APIKeys.Value,
		table.APIKeys.UserID,
		table.APIKeys.LastUsedAt,
		table.APIKeys.CreatedAt,
		table.APIKeys.UpdatedAt,
	).FROM(table.APIKeys).WHERE(
		table.APIKeys.ID.EQ(postgres.Int64(ID)),
	).LIMIT(1).Sql()

	key := model.APIKeys{}
	err := sqldb.
		QueryRow(ctx, query, args...).
		Scan(&key.ID, &key.Value, &key.UserID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not query api_key for ID")
		return nil, err
	}

	return &key, nil
}

// SaveApiKey saves the data of the key it used on. This function only saves
// the value and user ID from the struct and updates the timestamps. it
// saves the value as-is, always make sure to encrypt the API key before
// saving. SaveApiKey will update only the LastUserAt field when the key already exists.
func SaveApiKey(ctx context.Context, key *model.APIKeys) error {
	query, args := table.APIKeys.INSERT(
		table.APIKeys.Value,
		table.APIKeys.UserID,
		table.APIKeys.LastUsedAt,
	).VALUES(
		key.Value,
		key.UserID,
		key.LastUsedAt,
	).ON_CONFLICT().
		ON_CONSTRAINT("value_user_id_unique").
		DO_UPDATE(postgres.SET(table.APIKeys.LastUsedAt.SET(postgres.NOW()))).
		RETURNING(
			table.APIKeys.ID,
			table.APIKeys.LastUsedAt,
			table.APIKeys.UpdatedAt,
			table.APIKeys.CreatedAt,
		).Sql()

	log.Infof(query)
	err := sqldb.
		QueryRow(ctx, query, args...).
		Scan(&key.ID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not insert key (API key update is not supported)")
		return err
	}

	return nil
}

// DeleteApiKey deletes the API key is it called on.
func DeleteApiKey(ctx context.Context, key *model.APIKeys) error {
	query, args := table.APIKeys.
		DELETE().
		WHERE(table.APIKeys.ID.EQ(postgres.Int64(key.ID))).
		RETURNING(table.APIKeys.ID).
		Sql()

	deletedID := 0
	err := sqldb.QueryRow(ctx, query, args...).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Errorf("Could not delete key")
		return err
	}

	return nil
}
