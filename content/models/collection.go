package models

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

// Collection is a struct that mirrors the collection table
// from the database.
type Collection struct {
	ID        uint64
	Name      string
	UserID    uint64
	UpdatedAt time.Time
	CreatedAt time.Time
}

// NewCollection generates a new collection structure from a name and the
// associated user ID.
func NewCollection(name string, userID uint64) *Collection {
	return &Collection{
		Name:   name,
		UserID: userID,
	}
}

// ListCollections lists all collections for a given user, it returns
// a nil collection on an error.
func ListCollections(ctx context.Context, UserID uint64) ([]*Collection, error) {
	collectionQuery := `
		SELECT
			id,
			name,
			user_id,
			updated_at,
			created_at
		FROM
			"collections"
		WHERE
			collections.user_id = $1;
	`

	var collections []*Collection
	rows, err := sqldb.Query(ctx, collectionQuery, UserID)
	if err != nil {
		log.WithError(err).Error("Could not query collections")
		return nil, err
	}

	for rows.Next() {
		collection := &Collection{}

		err = rows.Scan(&collection.ID, &collection.Name, &collection.UserID, &collection.UpdatedAt, &collection.CreatedAt)
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
func GetCollectionByID(ctx context.Context, ID, UserID uint64) (*Collection, error) {
	collectionQuery := `
		SELECT
			id,
			name,
			user_id,
			updated_at,
			created_at
		FROM
			"collections"
		WHERE
			collections.id = $1 AND collections.user_id = $2
		LIMIT 1;
	`

	collection := Collection{}
	err := sqldb.
		QueryRow(ctx, collectionQuery, ID, UserID).
		Scan(&collection.ID, &collection.Name, &collection.UserID, &collection.UpdatedAt, &collection.CreatedAt)

	if err != nil {
		log.WithError(err).Errorf("Could not query collection for id %d", ID)
		return nil, err
	}

	return &collection, nil
}

// ValidateConstraint validates that no collection with the same name exists
// for a single user.
func (collection *Collection) ValidateConstraint(ctx context.Context) bool {
	collectionQuery := `
		SELECT
			id
		FROM
			"collections"
		WHERE
			collections.name = $1 AND collections.user_id = $2
		LIMIT 1;
	`

	id := 0
	err := sqldb.QueryRow(ctx, collectionQuery, collection.Name, collection.UserID).Scan(&id)
	if err == nil && id != 0 {
		log.Warning("Tried to save collection, a collection already exists for this name and user_id")
		return false
	}

	return true
}

// Save saves the data of the collection it used on. This method only saves
// the name and user ID from the struct and updates the timestamps. Save will
// trigger an error if the constraints are not respected.
func (collection *Collection) Save(ctx context.Context) error {
	if collection.ID == 0 {
		collectionQuery := `
		INSERT INTO "collections" (name, user_id)
		VALUES ($1, $2)
		RETURNING id, updated_at, created_at;
	`

		err := sqldb.
			QueryRow(ctx, collectionQuery, collection.Name, collection.UserID).
			Scan(&collection.ID, &collection.UpdatedAt, &collection.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert collection")
			return err
		}

		return nil
	}

	collectionQuery := `
		UPDATE "collections"
		SET name = $1, user_id = $2
		WHERE id = $3
		RETURNING id, updated_at, created_at;
	`

	err := sqldb.
		QueryRow(ctx, collectionQuery, collection.Name, collection.UserID, collection.ID).
		Scan(&collection.ID, &collection.UpdatedAt, &collection.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update collection")
		return err
	}

	return nil
}

// Delete deletes the Collection is it called on.
func (collection *Collection) Delete(ctx context.Context) error {
	collectionQuery := `
		DELETE FROM "collections"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, collectionQuery, collection.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Error("Could not delete collection")
		return err
	}

	return nil
}
