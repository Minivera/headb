package internal

import (
	"context"
	"encoding/json"

	"encore.app/content/helpers"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	log "github.com/sirupsen/logrus"

	"encore.app/content/convert"
	"encore.app/content/models"
	"encore.app/identity"
)

// ListDocuments lists all documents created by the authenticated user for a given collection
func ListDocuments(ctx context.Context, collectionID int64) ([]convert.DocumentPayload, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, collectionID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanReadDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	documents, err := models.ListDocuments(ctx, collection.ID)
	if err != nil {
		log.WithError(err).Error("Could not fetch documents for collection")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not fetch documents",
		}
	}

	payload, err := convert.DocumentModelsToPayloads(documents)
	if err != nil {
		log.WithError(err).Error("Could not convert documents to API safe version")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert documents for API",
		}
	}

	return payload, nil
}

// GetDocument finds a document by ID
func GetDocument(ctx context.Context, id int64) (convert.DocumentPayload, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, id, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	collection, err := helpers.GetCollection(ctx, document.CollectionID, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	if !helpers.CanReadDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert documents to API safe version")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return payload, nil
}

// CreateDocument creates a document for the authenticated user
func CreateDocument(ctx context.Context, collectionID int64, content json.RawMessage) (convert.DocumentPayload, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, collectionID, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	_, err = content.MarshalJSON()
	if string(content) == "null" || err != nil {
		log.WithError(err).Warning("Could not validate JSON on document request")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Received JSON string for content was not valid",
		}
	}

	document := models.NewDocument(string(content), collection.ID)

	err = models.SaveDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not save document")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert document to API safe version")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return payload, nil
}

// UpdateDocument updates a document by ID for the authenticated user
func UpdateDocument(ctx context.Context, id int64, content json.RawMessage) (convert.DocumentPayload, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, id, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	collection, err := helpers.GetCollection(ctx, document.CollectionID, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	_, err = content.MarshalJSON()
	if string(content) == "null" || err != nil {
		log.WithError(err).Warning("Could not validate JSON on document request")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Received JSON string for content was not valid",
		}
	}

	document.Content = string(content)

	err = models.SaveDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not save document")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert document to API safe version")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return payload, nil
}

// DeleteDocument deletes a document by ID for the authenticated user
func DeleteDocument(ctx context.Context, id int64) (convert.DocumentPayload, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, id, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	collection, err := helpers.GetCollection(ctx, document.CollectionID, userData.ID)
	if err != nil {
		return convert.DocumentPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	err = models.DeleteDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not delete document")
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		return convert.DocumentPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return payload, nil
}
