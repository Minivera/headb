package content

import (
	"context"

	log "github.com/sirupsen/logrus"

	"encore.app/content/convert"
	"encore.app/content/internal"
)

// ListCollectionsParams is the parameters for listing the collections of a database
type ListCollectionsParams struct {
	// The unique identifier of the database
	DatabaseID int64
}

// ListCollectionsResponse is the list of collections for the given database
type ListCollectionsResponse struct {
	// The fetched collections
	Collections []convert.CollectionPayload
}

// ListCollections lists all collections created by the authenticated user in the given
// database.
//encore:api auth
func ListCollections(ctx context.Context, params *ListCollectionsParams) (*ListCollectionsResponse, error) {
	collections, err := internal.ListCollections(ctx, params.DatabaseID)
	if err != nil {
		log.WithError(err).Warning("Could not fetch collections for this database")
		return nil, err
	}

	return &ListCollectionsResponse{
		Collections: collections,
	}, nil
}

// GetCollectionParams is the parameters for finding a collection by ID
type GetCollectionParams struct {
	// The unique identifier of the collection
	ID int64
}

// GetCollectionResponse is the result of having fetched a collection
type GetCollectionResponse struct {
	// The found collection
	Collection convert.CollectionPayload
}

// GetCollection Finds a collection by ID
//encore:api auth
func GetCollection(ctx context.Context, params *GetCollectionParams) (*GetCollectionResponse, error) {
	collection, err := internal.GetCollection(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &GetCollectionResponse{
		Collection: collection,
	}, nil
}

// CreateCollectionParams is the parameters for creating a collection for documents
type CreateCollectionParams struct {
	// The unique ID of the database to add this collection to
	DatabaseID int64

	// The name of the collection
	Name string
}

// CreateCollectionResponse is the result of creating a collection for documents
type CreateCollectionResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The created collection
	Collection convert.CollectionPayload
}

// CreateCollection creates a collection for the given database if owned by the authenticated user.
//encore:api auth
func CreateCollection(ctx context.Context, params *CreateCollectionParams) (*CreateCollectionResponse, error) {
	collection, err := internal.CreateCollection(ctx, params.DatabaseID, params.Name)
	if err != nil {
		return nil, err
	}

	return &CreateCollectionResponse{
		Message:    "Collection created successfully.",
		Collection: collection,
	}, nil
}

// UpdateCollectionParams is the parameters for updating a collection
type UpdateCollectionParams struct {
	// The unique identifier for the collection
	ID int64

	// The name of the collection
	Name string
}

// UpdateCollectionResponse is the result of updating a collection for documents
type UpdateCollectionResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The updated collection
	Collection convert.CollectionPayload
}

// UpdateCollection updates a collection by ID for the authenticated user
//encore:api auth
func UpdateCollection(ctx context.Context, params *UpdateCollectionParams) (*UpdateCollectionResponse, error) {
	collection, err := internal.UpdateCollection(ctx, params.ID, params.Name)
	if err != nil {
		return nil, err
	}

	return &UpdateCollectionResponse{
		Message:    "Collection updated successfully.",
		Collection: collection,
	}, nil
}

// DeleteCollectionParams is the parameters for deleting a collection
type DeleteCollectionParams struct {
	// The unique identifier for the collection
	ID int64
}

// DeleteCollectionResponse is the result of deleting a collection for documents
type DeleteCollectionResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The deleted collection
	Collection convert.CollectionPayload
}

// DeleteCollection deletes a collection by ID for the authenticated user
//encore:api auth
func DeleteCollection(ctx context.Context, params *DeleteCollectionParams) (*DeleteCollectionResponse, error) {
	collection, err := internal.DeleteCollection(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &DeleteCollectionResponse{
		Message:    "Collection deleted successfully.",
		Collection: collection,
	}, nil
}
