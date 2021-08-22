package models

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

// Document is a struct that mirrors the collection table
// from the database.
type Document struct {
	ID           uint64
	Content      map[string]interface{}
	CollectionID uint64
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

// NewDocument generates a new Document structure using the given content for
// a specific collection.
func NewDocument(content map[string]interface{}, collectionID uint64) *Document {
	return &Document{
		Content:      content,
		CollectionID: collectionID,
	}
}

// ListDocuments lists all documents for a given collection, returning an empty slice
// on an error.
func ListDocuments(ctx context.Context, CollectionID uint64) ([]*Document, error) {
	documentQuery := `
		SELECT
			id,
			content,
			collection_id,
			updated_at,
			created_at
		FROM
			"documents"
		WHERE
			documents.collection_id = $1;
	`

	var documents []*Document
	rows, err := sqldb.Query(ctx, documentQuery, CollectionID)
	if err != nil {
		log.WithError(err).Error("Could not query documents")
		return nil, err
	}

	for rows.Next() {
		document := &Document{}

		err = rows.Scan(&document.ID, &document.Content, &document.CollectionID, &document.UpdatedAt, &document.CreatedAt)
		if err != nil {
			log.WithError(err).Error("Could not scan documents")
			return nil, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}

// GetDocumentByUser fetches a single document record given an ID and the associated
// user ID of the collection this document belongs to. Returns nil on an error.
func GetDocumentByUser(ctx context.Context, ID, UserID uint64) (*Document, error) {
	documentQuery := `
		SELECT
			documents.id,
			content,
			collection_id,
			documents.updated_at,
			documents.created_at
		FROM
			"documents"
		LEFT JOIN "collections" ON documents.collection_id = collections.id AND collections.user_id = $2
		WHERE
			documents.id = $1
		LIMIT 1;
	`

	document := Document{}
	err := sqldb.
		QueryRow(ctx, documentQuery, ID, UserID).
		Scan(&document.ID, &document.Content, &document.CollectionID, &document.UpdatedAt, &document.CreatedAt)

	if err != nil {
		log.WithError(err).Errorf("Could not query document for id %v", ID)
		return nil, err
	}

	return &document, nil
}

// Save saves the data of the document it used on. This method only saves
// the content and collection ID from the struct and updates the timestamps. Save will
// trigger an error if the constraints are not respected.
func (document *Document) Save(ctx context.Context) error {
	if document.ID == 0 {
		documentQuery := `
		INSERT INTO "documents" (content, collection_id)
		VALUES ($1, $2)
		RETURNING id, updated_at, created_at;
	`

		err := sqldb.
			QueryRow(ctx, documentQuery, document.Content, document.CollectionID).
			Scan(&document.ID, &document.UpdatedAt, &document.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert document")
			return err
		}

		return nil
	}

	documentQuery := `
		UPDATE "documents"
		SET content = $1
		WHERE id = $2
		RETURNING id, updated_at, created_at;
	`

	err := sqldb.
		QueryRow(ctx, documentQuery, document.Content, document.ID).
		Scan(&document.ID, &document.UpdatedAt, &document.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update document")
		return err
	}

	return nil
}

// Delete deletes the Document is it called on.
func (document *Document) Delete(ctx context.Context) error {
	documentQuery := `
		DELETE FROM "documents"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, documentQuery, document.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Error("Could not delete document")
		return err
	}

	return nil
}
