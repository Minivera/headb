package content

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

// ListDocumentsParams is the parameters for listing the documents of a collection
type ListDocumentsParams struct {
	// The unique identifier of the collection
	CollectionID int64
}

// ListDocumentsResponse is the list of documents for the current user and identified collection
type ListDocumentsResponse struct {
	// The fetched collections
	Documents []convert.DocumentPayload
}

// ListDocuments lists all documents created by the authenticated user for a given collection
//encore:api auth
func ListDocuments(ctx context.Context, params *ListDocumentsParams) (*ListDocumentsResponse, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, params.CollectionID, userData.ID)
	if err != nil {
		return nil, err
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

	return &ListDocumentsResponse{
		Documents: payload,
	}, nil
}

// GetDocumentParams is the parameters for finding a document by ID
type GetDocumentParams struct {
	// The unique identifier of the document
	ID int64
}

// GetDocumentResponse is the result of having fetched a document
type GetDocumentResponse struct {
	// The found document
	Document convert.DocumentPayload
}

// GetDocument finds a document by ID
//encore:api auth
func GetDocument(ctx context.Context, params *GetDocumentParams) (*GetDocumentResponse, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert documents to API safe version")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return &GetDocumentResponse{
		Document: payload,
	}, nil
}

// CreateDocumentParams is the parameters for creating a document in a collection
type CreateDocumentParams struct {
	// The unique identifier for the collection this document should be added to
	CollectionID int64

	// The content of the document
	Content json.RawMessage
}

// CreateDocumentResponse is the result of creating a document
type CreateDocumentResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The created document
	Document convert.DocumentPayload
}

// CreateDocument creates a document for the authenticated user
//encore:api auth
func CreateDocument(ctx context.Context, params *CreateDocumentParams) (*CreateDocumentResponse, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, params.CollectionID, userData.ID)
	if err != nil {
		return nil, err
	}

	_, err = params.Content.MarshalJSON()
	if string(params.Content) == "null" || err != nil {
		log.WithError(err).Warning("Could not validate JSON on document request")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Received JSON string for content was not valid",
		}
	}

	document := models.NewDocument(string(params.Content), collection.ID)

	err = models.SaveDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not save document")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert document to API safe version")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return &CreateDocumentResponse{
		Message:  "Document created successfully.",
		Document: payload,
	}, nil
}

// UpdateDocumentParams is the parameters for updating a document
type UpdateDocumentParams struct {
	// The unique identifier for the document
	ID int64

	// The content of the document
	Content json.RawMessage
}

// UpdateDocumentResponse is the result of updating a document for documents
type UpdateDocumentResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The updated document
	Document convert.DocumentPayload
}

// UpdateDocument updates a document by ID for the authenticated user
//encore:api auth
func UpdateDocument(ctx context.Context, params *UpdateDocumentParams) (*UpdateDocumentResponse, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	_, err = params.Content.MarshalJSON()
	if string(params.Content) == "null" || err != nil {
		log.WithError(err).Warning("Could not validate JSON on document request")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Received JSON string for content was not valid",
		}
	}

	document.Content = string(params.Content)

	err = models.SaveDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not save document")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		log.WithError(err).Error("Could not convert document to API safe version")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return &UpdateDocumentResponse{
		Message:  "Document updated successfully.",
		Document: payload,
	}, nil
}

// DeleteDocumentParams is the parameters for deleting a document
type DeleteDocumentParams struct {
	// The unique identifier for the document
	ID int64
}

// DeleteDocumentResponse is the result of deleting a document for documents
type DeleteDocumentResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The deleted document
	Document convert.DocumentPayload
}

// DeleteDocument deletes a document by ID for the authenticated user
//encore:api auth
func DeleteDocument(ctx context.Context, params *DeleteDocumentParams) (*DeleteDocumentResponse, error) {
	userData := auth.Data().(*identity.UserData)

	document, err := helpers.GetDocument(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	err = models.DeleteDocument(ctx, document)
	if err != nil {
		log.WithError(err).Error("Could not delete document")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete document",
		}
	}

	payload, err := convert.DocumentModelToPayload(document)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not convert document for API",
		}
	}

	return &DeleteDocumentResponse{
		Message:  "Document deleted successfully.",
		Document: payload,
	}, nil
}
