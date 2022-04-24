package models

import (
	"context"

	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
)

var db = sqldb.Named("identity").Stdlib()

// NewApiKey generates a new API key structure from an encrypted key
// value and a user ID. Never give plain API key values to this function.
func NewApiKey(value string, userID uuid.UUID) *model.APIKeys {
	return &model.APIKeys{
		Value:  value,
		UserID: userID,
	}
}

// ListApiKeysForUser fetches all the API keys for a specific user, returns an empty array
// on an error.
func ListApiKeysForUser(ctx context.Context, userID uuid.UUID) ([]*model.APIKeys, error) {
	statement := postgres.SELECT(
		table.APIKeys.ID,
		table.APIKeys.Value,
		table.APIKeys.UserID,
		table.APIKeys.LastUsedAt,
		table.APIKeys.CreatedAt,
		table.APIKeys.UpdatedAt,
	).FROM(table.APIKeys).WHERE(
		table.APIKeys.UserID.EQ(postgres.UUID(userID)),
	)

	var keys []*model.APIKeys
	err := statement.QueryContext(ctx, db, &keys)
	if err != nil {
		log.WithError(err).Error("Could not query api keys")
		return nil, err
	}

	return keys, nil
}

// GetApiKey gets an API key from a database ID, returning it
// if successfully fetched. This should only be called internally as
// it risks exposing keys not owned by users if they guess the ID.
func GetApiKey(ctx context.Context, id uuid.UUID) (*model.APIKeys, error) {
	statement := postgres.SELECT(
		table.APIKeys.ID,
		table.APIKeys.Value,
		table.APIKeys.UserID,
		table.APIKeys.LastUsedAt,
		table.APIKeys.CreatedAt,
		table.APIKeys.UpdatedAt,
	).FROM(table.APIKeys).WHERE(
		table.APIKeys.ID.EQ(postgres.UUID(id)),
	).LIMIT(1)

	key := model.APIKeys{}
	err := statement.QueryContext(ctx, db, &key)
	if err != nil {
		log.WithError(err).Error("Could not query api_key for ID")
		return nil, err
	}

	return &key, nil
}

// GetApiKeyForUser gets an API key from a database ID for a specific user, returning it
// if successfully fetched.
func GetApiKeyForUser(ctx context.Context, id, userID uuid.UUID) (*model.APIKeys, error) {
	statement := postgres.SELECT(
		table.APIKeys.ID,
		table.APIKeys.Value,
		table.APIKeys.UserID,
		table.APIKeys.LastUsedAt,
		table.APIKeys.CreatedAt,
		table.APIKeys.UpdatedAt,
	).FROM(table.APIKeys).WHERE(
		table.APIKeys.ID.EQ(postgres.UUID(id)).AND(table.APIKeys.UserID.EQ(postgres.UUID(userID))),
	).LIMIT(1)

	key := model.APIKeys{}
	err := statement.QueryContext(ctx, db, &key)
	if err != nil {
		log.WithError(err).Error("Could not query api_key for ID and userID")
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

	err := db.
		QueryRowContext(ctx, query, args...).
		Scan(&key.ID, &key.LastUsedAt, &key.UpdatedAt, &key.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not insert key (API key update is not supported)")
		return err
	}

	return nil
}

// TransferApiKeys will transfer all api keys from a user to another. This is useful when
// deleting temporary users to merge their data with an existing user. We should NEVER
// allow transfer of API keys in other contexts.
func TransferApiKeys(ctx context.Context, oldUserID, newUserID uuid.UUID) error {
	query, args := table.APIKeys.UPDATE().SET(
		table.APIKeys.UserID.SET(postgres.UUID(newUserID)),
	).WHERE(table.APIKeys.UserID.EQ(postgres.UUID(oldUserID))).Sql()

	transferredCount := 0
	err := db.QueryRowContext(ctx, query, args...).Scan(&transferredCount)

	if err != nil {
		log.WithError(err).Errorf("Could not delete key")
		return err
	}

	log.WithFields(map[string]interface{}{
		"old_user_id": oldUserID,
		"new_user_id": newUserID,
		"count":       transferredCount,
	}).Info("Transferred API keys from old user to new user.")

	return nil
}

// DeleteApiKey deletes the API key is it called on.
func DeleteApiKey(ctx context.Context, key *model.APIKeys) error {
	query, args := table.APIKeys.
		DELETE().
		WHERE(table.APIKeys.ID.EQ(postgres.UUID(key.ID))).
		RETURNING(table.APIKeys.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)

	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Errorf("Could not delete key")
		return err
	}

	return nil
}
