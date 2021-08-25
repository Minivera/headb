package models

import (
	"context"

	"encore.dev/storage/sqldb"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

// NewCollection generates a new collection structure from a name and the
// associated user ID.
func NewCollection(name string, userID int64) *model.Collections {
	return &model.Collections{
		Name:   name,
		UserID: userID,
	}
}

// ListCollections lists all collections for a given user, it returns
// a nil collection on an error.
func ListCollections(ctx context.Context, userID int64) ([]*model.Collections, error) {
	query, args := postgres.SELECT(
		table.Collections.ID,
		table.Collections.Name,
		table.Collections.UserID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).FROM(table.Collections).WHERE(
		table.Collections.UserID.EQ(postgres.Int64(userID)),
	).Sql()

	var collections []*model.Collections
	rows, err := sqldb.Query(ctx, query, args...)
	if err != nil {
		log.WithError(err).Error("Could not query collections")
		return nil, err
	}

	for rows.Next() {
		collection := &model.Collections{}

		err = rows.Scan(
			&collection.ID,
			&collection.Name,
			&collection.UserID,
			&collection.UpdatedAt,
			&collection.CreatedAt,
		)
		if err != nil {
			log.WithError(err).Error("Could not scan collections")
			return nil, err
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

// GetCollectionByID fetches a single collection record given an ID and the associated
// user ID. Returns nil on an error.
func GetCollectionByID(ctx context.Context, id, userID int64) (*model.Collections, error) {
	query, args := postgres.SELECT(
		table.Collections.ID,
		table.Collections.Name,
		table.Collections.UserID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).FROM(
		table.Collections,
	).WHERE(
		table.Collections.ID.EQ(postgres.Int64(id)).
			AND(table.Collections.UserID.EQ(postgres.Int64(userID))),
	).LIMIT(1).Sql()

	collection := model.Collections{}
	err := sqldb.
		QueryRow(ctx, query, args...).
		Scan(
			&collection.ID,
			&collection.Name,
			&collection.UserID,
			&collection.UpdatedAt,
			&collection.CreatedAt,
		)

	if err != nil {
		log.WithError(err).Errorf("Could not query collection for id %d", id)
		return nil, err
	}

	return &collection, nil
}

// ValidateCollectionConstraint validates that no collection with the same name exists
// for a single user.
func ValidateCollectionConstraint(ctx context.Context, collection *model.Collections) bool {
	query, args := postgres.SELECT(
		table.Collections.ID,
	).FROM(
		table.Collections,
	).WHERE(
		table.Collections.Name.EQ(postgres.String(collection.Name)).
			AND(table.Collections.UserID.EQ(postgres.Int64(collection.UserID))),
	).LIMIT(1).Sql()

	id := 0
	err := sqldb.QueryRow(ctx, query, args...).Scan(&id)
	if err == nil && id != 0 {
		log.Warning("Tried to save collection, a collection already exists for this name and user_id")
		return false
	}

	return true
}

// SaveCollection saves the data of the collection it used on. This method only saves
// the name and user ID from the struct and updates the timestamps. SaveCollection will
// trigger an error if the constraints are not respected.
func SaveCollection(ctx context.Context, collection *model.Collections) error {
	if collection.ID == 0 {
		query, args := table.Collections.INSERT(
			table.Collections.Name,
			table.Collections.UserID,
		).VALUES(
			collection.Name,
			collection.UserID,
		).RETURNING(
			table.Collections.ID,
			table.Collections.UpdatedAt,
			table.Collections.CreatedAt,
		).Sql()

		err := sqldb.
			QueryRow(ctx, query, args...).
			Scan(&collection.ID, &collection.UpdatedAt, &collection.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert collection")
			return err
		}

		return nil
	}

	query, args := table.Collections.UPDATE().SET(
		table.Collections.Name.SET(postgres.String(collection.Name)),
		table.Collections.UserID.SET(postgres.Int64(collection.UserID)),
	).WHERE(
		table.Collections.ID.EQ(postgres.Int64(collection.ID)),
	).RETURNING(
		table.Collections.ID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).Sql()

	err := sqldb.
		QueryRow(ctx, query, args...).
		Scan(&collection.ID, &collection.UpdatedAt, &collection.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update collection")
		return err
	}

	return nil
}

// DeleteCollection deletes the Collection is it called on.
func DeleteCollection(ctx context.Context, collection *model.Collections) error {
	query, args := table.Collections.
		DELETE().
		WHERE(table.Collections.ID.EQ(postgres.Int64(collection.ID))).
		RETURNING(table.Collections.ID).
		Sql()

	deletedID := 0
	err := sqldb.QueryRow(ctx, query, args...).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Error("Could not delete collection")
		return err
	}

	return nil
}
