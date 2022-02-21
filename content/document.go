package content

import (
	"context"
	"encoding/json"

	"encore.app/content/convert"
	"encore.app/content/internal"
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
	documents, err := internal.ListDocuments(ctx, params.CollectionID)
	if err != nil {
		return nil, err
	}

	return &ListDocumentsResponse{
		Documents: documents,
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
	document, err := internal.GetDocument(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &GetDocumentResponse{
		Document: document,
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
	document, err := internal.CreateDocument(ctx, params.CollectionID, params.Content)
	if err != nil {
		return nil, err
	}

	return &CreateDocumentResponse{
		Message:  "Document created successfully.",
		Document: document,
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
	document, err := internal.UpdateDocument(ctx, params.ID, params.Content)
	if err != nil {
		return nil, err
	}

	return &UpdateDocumentResponse{
		Message:  "Document updated successfully.",
		Document: document,
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
	document, err := internal.DeleteDocument(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &DeleteDocumentResponse{
		Message:  "Document deleted successfully.",
		Document: document,
	}, nil
}
