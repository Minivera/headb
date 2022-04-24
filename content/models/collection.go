package models

import (
	"context"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

// NewCollection generates a new collection structure from a name and the
// associated database ID.
func NewCollection(name string, databaseID uuid.UUID) *model.Collections {
	return &model.Collections{
		Name:       name,
		DatabaseID: databaseID,
	}
}

// ListCollections lists all collections for a given database, it returns
// a nil collection on an error.
func ListCollections(ctx context.Context, databaseID uuid.UUID) ([]*model.Collections, error) {
	statement := postgres.SELECT(
		table.Collections.ID,
		table.Collections.Name,
		table.Collections.DatabaseID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).FROM(table.Collections).WHERE(
		table.Collections.DatabaseID.EQ(postgres.UUID(databaseID)),
	)

	var collections []*model.Collections
	err := statement.QueryContext(ctx, db, &collections)
	if err != nil {
		log.WithError(err).Error("Could not query collections")
		return nil, err
	}

	return collections, nil
}

// GetCollectionByID fetches a single collection record given an ID and the associated
// database ID. Returns nil on an error.
func GetCollectionByID(ctx context.Context, id, userID uuid.UUID) (*model.Collections, error) {
	statement := postgres.SELECT(
		table.Collections.ID,
		table.Collections.Name,
		table.Collections.DatabaseID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).FROM(
		table.Collections.LEFT_JOIN(
			table.Databases,
			table.Collections.DatabaseID.EQ(table.Databases.ID),
		),
	).WHERE(
		table.Collections.ID.EQ(postgres.UUID(id)).
			AND(table.Databases.UserID.EQ(postgres.UUID(userID))),
	).LIMIT(1)

	collection := model.Collections{}
	err := statement.QueryContext(ctx, db, &collection)
	if err != nil {
		log.WithError(err).Errorf("Could not query collection for id %d", id)
		return nil, err
	}

	return &collection, nil
}

// ValidateCollectionConstraint validates that no collection with the same name exists
// for a single database.
func ValidateCollectionConstraint(ctx context.Context, collection *model.Collections) bool {
	query, args := postgres.SELECT(
		table.Collections.ID,
	).FROM(
		table.Collections,
	).WHERE(
		table.Collections.Name.EQ(postgres.String(collection.Name)).
			AND(table.Collections.DatabaseID.EQ(postgres.UUID(collection.DatabaseID))),
	).LIMIT(1).Sql()

	id := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == nil && id != uuid.Nil {
		log.Warning("Tried to save collection, a collection already exists for this name and database_id")
		return false
	}

	return true
}

// SaveCollection saves the data of the collection it used on. This method only saves
// the name and database ID from the struct and updates the timestamps. SaveCollection will
// trigger an error if the constraints are not respected.
func SaveCollection(ctx context.Context, collection *model.Collections) error {
	if collection.ID == uuid.Nil {
		query, args := table.Collections.INSERT(
			table.Collections.Name,
			table.Collections.DatabaseID,
		).VALUES(
			collection.Name,
			collection.DatabaseID,
		).RETURNING(
			table.Collections.ID,
			table.Collections.UpdatedAt,
			table.Collections.CreatedAt,
		).Sql()

		err := db.
			QueryRowContext(ctx, query, args...).
			Scan(&collection.ID, &collection.UpdatedAt, &collection.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert collection")
			return err
		}

		return nil
	}

	query, args := table.Collections.UPDATE().SET(
		table.Collections.Name.SET(postgres.String(collection.Name)),
		table.Collections.DatabaseID.SET(postgres.UUID(collection.DatabaseID)),
	).WHERE(
		table.Collections.ID.EQ(postgres.UUID(collection.ID)),
	).RETURNING(
		table.Collections.ID,
		table.Collections.UpdatedAt,
		table.Collections.CreatedAt,
	).Sql()

	err := db.
		QueryRowContext(ctx, query, args...).
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
		WHERE(table.Collections.ID.EQ(postgres.UUID(collection.ID))).
		RETURNING(table.Collections.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete collection")
		return err
	}

	return nil
}
