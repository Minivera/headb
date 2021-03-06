package models

import (
	"context"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

// NewDocument generates a new Document structure using the given content for
// a specific collection.
func NewDocument(content string, collectionID uuid.UUID) *model.Documents {
	return &model.Documents{
		Content:      content,
		CollectionID: collectionID,
	}
}

// ListDocuments lists all documents for a given collection, returning an empty slice
// on an error.
func ListDocuments(ctx context.Context, CollectionID uuid.UUID) ([]*model.Documents, error) {
	statement := postgres.SELECT(
		table.Documents.ID,
		table.Documents.Content,
		table.Documents.CollectionID,
		table.Documents.UpdatedAt,
		table.Documents.CreatedAt,
	).FROM(table.Documents).WHERE(
		table.Documents.CollectionID.EQ(postgres.UUID(CollectionID)),
	)

	var documents []*model.Documents
	err := statement.QueryContext(ctx, db, &documents)
	if err != nil {
		log.WithError(err).Error("Could not query documents")
		return nil, err
	}

	return documents, nil
}

// GetDocumentByUser fetches a single document record given an ID and the associated
// user ID of the collection this document belongs to. Returns nil on an error.
func GetDocumentByUser(ctx context.Context, ID, UserID uuid.UUID) (*model.Documents, error) {
	statement := postgres.SELECT(
		table.Documents.ID,
		table.Documents.Content,
		table.Documents.CollectionID,
		table.Documents.UpdatedAt,
		table.Documents.CreatedAt,
	).FROM(
		table.Documents.LEFT_JOIN(
			table.Collections,
			table.Documents.CollectionID.EQ(table.Collections.ID),
		).LEFT_JOIN(
			table.Databases,
			table.Collections.DatabaseID.EQ(table.Databases.ID),
		),
	).WHERE(
		table.Documents.ID.EQ(postgres.UUID(ID)).
			AND(table.Databases.UserID.EQ(postgres.UUID(UserID))),
	).LIMIT(1)

	document := model.Documents{}
	err := statement.QueryContext(ctx, db, &document)
	if err != nil {
		log.WithError(err).Errorf("Could not query document for id %v", ID)
		return nil, err
	}

	return &document, nil
}

// SaveDocument saves the data of the document it used on. This method only saves
// the content and collection ID from the struct and updates the timestamps. SaveDocument will
// trigger an error if the constraints are not respected.
func SaveDocument(ctx context.Context, document *model.Documents) error {
	if document.ID == uuid.Nil {
		query, args := table.Documents.INSERT(
			table.Documents.Content,
			table.Documents.CollectionID,
		).VALUES(
			document.Content,
			document.CollectionID,
		).RETURNING(
			table.Documents.ID,
			table.Documents.UpdatedAt,
			table.Documents.CreatedAt,
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

	query, args := table.Documents.UPDATE().SET(
		table.Documents.Content.SET(postgres.String(document.Content)),
	).WHERE(
		table.Documents.ID.EQ(postgres.UUID(document.ID)),
	).RETURNING(
		table.Documents.ID,
		table.Documents.UpdatedAt,
		table.Documents.CreatedAt,
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

// DeleteDocument deletes the Document is it called on.
func DeleteDocument(ctx context.Context, document *model.Documents) error {
	query, args := table.Documents.
		DELETE().
		WHERE(table.Documents.ID.EQ(postgres.UUID(document.ID))).
		RETURNING(table.Documents.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete document")
		return err
	}

	return nil
}
