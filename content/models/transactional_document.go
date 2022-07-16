package models

import (
	"context"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

// NewTransactionalDocument generates a new transactional document structure using the given content for
// a specific collection.
func NewTransactionalDocument(content string, collectionID, transactionID uuid.UUID) *model.TransactionalDocuments {
	return &model.TransactionalDocuments{
		Content:       content,
		CollectionID:  collectionID,
		TransactionID: transactionID,
	}
}

// ListTransactionalDocuments lists all transactional documents for a given collection, returning an empty slice
// on an error.
func ListTransactionalDocuments(ctx context.Context, collectionID, transactionID uuid.UUID) ([]*model.TransactionalDocuments, error) {
	statement := postgres.SELECT(
		table.TransactionalDocuments.ID,
		table.TransactionalDocuments.Content,
		table.TransactionalDocuments.CollectionID,
		table.TransactionalDocuments.TransactionID,
		table.TransactionalDocuments.UpdatedAt,
		table.TransactionalDocuments.CreatedAt,
	).FROM(table.TransactionalDocuments).WHERE(
		table.TransactionalDocuments.CollectionID.EQ(postgres.UUID(collectionID)).
			AND(table.TransactionalDocuments.TransactionID.EQ(postgres.UUID(transactionID))),
	)

	var documents []*model.TransactionalDocuments
	err := statement.QueryContext(ctx, db, &documents)
	if err != nil {
		log.WithError(err).Error("Could not query documents")
		return nil, err
	}

	return documents, nil
}

// GetTransactionalDocumentByUser fetches a single transactional document record given an ID and the associated
// user ID of the collection this document belongs to. Returns nil on an error.
func GetTransactionalDocumentByUser(ctx context.Context, id, userID, transactionID uuid.UUID) (*model.TransactionalDocuments, error) {
	statement := postgres.SELECT(
		table.TransactionalDocuments.ID,
		table.TransactionalDocuments.Content,
		table.TransactionalDocuments.CollectionID,
		table.TransactionalDocuments.TransactionID,
		table.TransactionalDocuments.UpdatedAt,
		table.TransactionalDocuments.CreatedAt,
	).FROM(
		table.TransactionalDocuments.LEFT_JOIN(
			table.Collections,
			table.TransactionalDocuments.CollectionID.EQ(table.Collections.ID),
		).LEFT_JOIN(
			table.Databases,
			table.Collections.DatabaseID.EQ(table.Databases.ID),
		),
	).WHERE(
		table.TransactionalDocuments.ID.EQ(postgres.UUID(id)).
			AND(table.TransactionalDocuments.TransactionID.EQ(postgres.UUID(transactionID))).
			AND(table.Databases.UserID.EQ(postgres.UUID(userID))),
	).LIMIT(1)

	document := model.TransactionalDocuments{}
	err := statement.QueryContext(ctx, db, &document)
	if err != nil {
		log.WithError(err).Errorf("Could not query document for id %v", id)
		return nil, err
	}

	return &document, nil
}

// SaveTransactionalDocument saves the data of the transactional document it used on. This method only saves
// the content and collection ID from the struct and updates the timestamps. SaveTransactionalDocument will
// trigger an error if the constraints are not respected.
func SaveTransactionalDocument(ctx context.Context, document *model.TransactionalDocuments) error {
	if document.ID == uuid.Nil {
		query, args := table.Documents.INSERT(
			table.TransactionalDocuments.Content,
			table.TransactionalDocuments.CollectionID,
			table.TransactionalDocuments.TransactionID,
		).VALUES(
			document.Content,
			document.CollectionID,
			document.TransactionID,
		).RETURNING(
			table.TransactionalDocuments.ID,
			table.TransactionalDocuments.UpdatedAt,
			table.TransactionalDocuments.CreatedAt,
		).Sql()

		err := db.
			QueryRowContext(ctx, query, args...).
			Scan(&document.ID, &document.UpdatedAt, &document.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert document")
			return err
		}

		return nil
	}

	// TODO: Consider we might have to create the document first if updating a document
	// that exists outside the transaction
	query, args := table.TransactionalDocuments.UPDATE().SET(
		table.TransactionalDocuments.Content.SET(postgres.String(document.Content)),
	).WHERE(
		table.TransactionalDocuments.ID.EQ(postgres.UUID(document.ID)),
	).RETURNING(
		table.TransactionalDocuments.ID,
		table.TransactionalDocuments.UpdatedAt,
		table.TransactionalDocuments.CreatedAt,
	).Sql()

	err := db.
		QueryRowContext(ctx, query, args...).
		Scan(&document.ID, &document.UpdatedAt, &document.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update document")
		return err
	}

	return nil
}

// DeleteTransactionalDocument deletes the transactional document is it called on.
func DeleteTransactionalDocument(ctx context.Context, document *model.TransactionalDocuments) error {
	query, args := table.TransactionalDocuments.
		DELETE().
		WHERE(table.TransactionalDocuments.ID.EQ(postgres.UUID(document.ID))).
		RETURNING(table.TransactionalDocuments.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete document")
		return err
	}

	return nil
}
