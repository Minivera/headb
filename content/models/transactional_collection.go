package models

import (
	"context"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

// NewTransactionalCollection generates a new transaction collection structure from a name and the
// associated database ID.
func NewTransactionalCollection(name string, databaseID, transactionID uuid.UUID) *model.TransactionalCollections {
	return &model.TransactionalCollections{
		Name:          name,
		DatabaseID:    databaseID,
		TransactionID: transactionID,
	}
}

// ListTransactionalCollections lists all collections for a given database in a transaction, it returns
// a nil collection on an error.
func ListTransactionalCollections(ctx context.Context, databaseID, transactionID uuid.UUID) ([]*model.TransactionalCollections, error) {
	statement := postgres.SELECT(
		table.TransactionalCollections.ID,
		table.TransactionalCollections.Name,
		table.TransactionalCollections.DatabaseID,
		table.TransactionalCollections.TransactionID,
		table.TransactionalCollections.UpdatedAt,
		table.TransactionalCollections.CreatedAt,
	).FROM(table.TransactionalCollections).WHERE(
		table.TransactionalCollections.DatabaseID.EQ(postgres.UUID(databaseID)).AND(
			table.TransactionalCollections.TransactionID.EQ(postgres.UUID(transactionID)),
		),
	)

	var collections []*model.TransactionalCollections
	err := statement.QueryContext(ctx, db, &collections)
	if err != nil {
		log.WithError(err).Error("Could not query collections")
		return nil, err
	}

	return collections, nil
}

// GetTransactionalCollectionByID fetches a single transactional collection record given an ID, and the associated
// database ID and transaction ID. Returns nil on an error.
func GetTransactionalCollectionByID(ctx context.Context, id, userID, transactionID uuid.UUID) (*model.TransactionalCollections, error) {
	statement := postgres.SELECT(
		table.TransactionalCollections.ID,
		table.TransactionalCollections.Name,
		table.TransactionalCollections.DatabaseID,
		table.TransactionalCollections.TransactionID,
		table.TransactionalCollections.UpdatedAt,
		table.TransactionalCollections.CreatedAt,
	).FROM(
		table.TransactionalCollections.LEFT_JOIN(
			table.Databases,
			table.TransactionalCollections.DatabaseID.EQ(table.Databases.ID),
		),
	).WHERE(
		table.TransactionalCollections.ID.EQ(postgres.UUID(id)).
			AND(table.TransactionalCollections.TransactionID.EQ(postgres.UUID(transactionID))).
			AND(table.Databases.UserID.EQ(postgres.UUID(userID))),
	).LIMIT(1)

	collection := model.TransactionalCollections{}
	err := statement.QueryContext(ctx, db, &collection)
	if err != nil {
		log.WithError(err).Errorf("Could not query collection for id %d", id)
		return nil, err
	}

	return &collection, nil
}

// ValidateTransactionalCollectionConstraint validates that no collection nor transactional with the same name exists
// for a single database.
func ValidateTransactionalCollectionConstraint(ctx context.Context, collection *model.TransactionalCollections) bool {
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

	query, args = postgres.SELECT(
		table.TransactionalCollections.ID,
	).FROM(
		table.TransactionalCollections,
	).WHERE(
		table.TransactionalCollections.Name.EQ(postgres.String(collection.Name)).
			AND(table.TransactionalCollections.DatabaseID.EQ(postgres.UUID(collection.DatabaseID))),
	).LIMIT(1).Sql()

	id = uuid.Nil
	err = db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == nil && id != uuid.Nil {
		log.Warning("Tried to save collection, a collection already exists for this name and database_id")
		return false
	}

	return true
}

// SaveTransactionalCollection saves the data of the transactional collection it used on. This method only saves
// the name and database ID from the struct and updates the timestamps. SaveTransactionalCollection will
// trigger an error if the constraints are not respected.
func SaveTransactionalCollection(ctx context.Context, collection *model.TransactionalCollections) error {
	if collection.ID == uuid.Nil {
		query, args := table.Collections.INSERT(
			table.TransactionalCollections.Name,
			table.TransactionalCollections.DatabaseID,
			table.TransactionalCollections.TransactionID,
		).VALUES(
			collection.Name,
			collection.DatabaseID,
			collection.TransactionID,
		).RETURNING(
			table.TransactionalCollections.ID,
			table.TransactionalCollections.UpdatedAt,
			table.TransactionalCollections.CreatedAt,
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

	// TODO: Consider we might have to create the collection first if updating a collection
	// that exists outside the transaction
	query, args := table.TransactionalCollections.UPDATE().SET(
		table.TransactionalCollections.Name.SET(postgres.String(collection.Name)),
		table.TransactionalCollections.DatabaseID.SET(postgres.UUID(collection.DatabaseID)),
		table.TransactionalCollections.TransactionID.SET(postgres.UUID(collection.TransactionID)),
	).WHERE(
		table.TransactionalCollections.ID.EQ(postgres.UUID(collection.ID)),
	).RETURNING(
		table.TransactionalCollections.ID,
		table.TransactionalCollections.UpdatedAt,
		table.TransactionalCollections.CreatedAt,
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

// DeleteTransactionalCollection deletes the transactional collection is it called on.
func DeleteTransactionalCollection(ctx context.Context, collection *model.TransactionalCollections) error {
	query, args := table.TransactionalCollections.
		DELETE().
		WHERE(table.TransactionalCollections.ID.EQ(postgres.UUID(collection.ID))).
		RETURNING(table.TransactionalCollections.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete collection")
		return err
	}

	return nil
}
