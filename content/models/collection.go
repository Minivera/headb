package models

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

type Collection struct {
	ID        uint64
	Name      string
	UserID    uint64
	UpdatedAt time.Time
	CreatedAt time.Time
}

func NewCollection(name string, userID uint64) *Collection {
	return &Collection{
		Name:   name,
		UserID: userID,
	}
}

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
		log.Errorf("Could not query collections, %v", err)
		return nil, err
	}

	for rows.Next() {
		collection := &Collection{}

		err = rows.Scan(&collection.ID, &collection.Name, &collection.UserID, &collection.UpdatedAt, &collection.CreatedAt)
		if err != nil {
			log.Errorf("Could not scan collections, %v", err)
			return nil, err
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

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
		log.Errorf("Could not query collection for id %v, %v", ID, err)
		return nil, err
	}

	return &collection, nil
}

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
		log.Errorf("Tried to save collection, a collection already exists for this name and user_id")
		return false
	}

	return true
}

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
			log.Errorf("Could not insert collection, %v", err)
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
		log.Errorf("Could not update collection, %v", err)
		return err
	}

	return nil
}

func (collection *Collection) Delete(ctx context.Context) error {
	collectionQuery := `
		DELETE FROM "collections"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, collectionQuery, collection.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.Errorf("Could not delete collection, %v", err)
		return err
	}

	return nil
}
