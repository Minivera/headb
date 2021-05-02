package headb

import (
	"context"
	"errors"
	"fmt"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"

	"encore.app/headb/convert"
	"encore.app/headb/models"
)

// ListCollectionsResponse Is the list of collections for the current user
type ListCollectionsResponse struct {
	// The fetched collections
	Collections []convert.CollectionPayload
}

// ListCollections lists all collections created by the authenticated user
//encore:api auth
func ListCollections(ctx context.Context) (*ListCollectionsResponse, error) {
	userData := auth.Data().(*UserData)

	collections, err := models.ListCollections(ctx, userData.ID)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not fetch collections",
		}
	}

	return &ListCollectionsResponse{
		Collections: convert.CollectionModelsToPayloads(collections),
	}, nil
}

// GetCollectionParams is the parameters for finding a collection by ID
type GetCollectionParams struct {
	// The unique identifier of the collection
	ID uint64
}

// GetCollectionResponse is the result of having fetched a collection
type GetCollectionResponse struct {
	// The found collection
	Collection convert.CollectionPayload
}

// GetCollection Finds a collection by ID
//encore:api auth
func GetCollection(ctx context.Context, params *GetCollectionParams) (*GetCollectionResponse, error) {
	userData := auth.Data().(*UserData)

	collection, err := models.GetCollectionByID(ctx, params.ID, userData.ID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find collection",
		}
	} else if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find collection, unknown error",
		}
	}

	return &GetCollectionResponse{
		Collection: convert.CollectionModelToPayload(collection),
	}, nil
}

// CreateCollectionParams is the parameters for creating a collection for documents
type CreateCollectionParams struct {
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

// CreateCollection creates a collection for the authenticated user
//encore:api auth
func CreateCollection(ctx context.Context, params *CreateCollectionParams) (*CreateCollectionResponse, error) {
	userData := auth.Data().(*UserData)

	collection := models.NewCollection(params.Name, userData.ID)
	if !collection.ValidateConstraint(ctx) {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists for this user", collection.Name),
		}
	}

	err := collection.Save(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save collection",
		}
	}

	return &CreateCollectionResponse{
		Message:    "Collection created successfully.",
		Collection: convert.CollectionModelToPayload(collection),
	}, nil
}

// UpdateCollectionParams is the parameters for updating a collection
type UpdateCollectionParams struct {
	// The unique identifier for the collection
	ID uint64

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
	userData := auth.Data().(*UserData)

	collection, err := models.GetCollectionByID(ctx, params.ID, userData.ID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find collection",
		}
	} else if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find collection, unknown error",
		}
	}

	collection.Name = params.Name
	if !collection.ValidateConstraint(ctx) {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists for this user", collection.Name),
		}
	}

	err = collection.Save(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save collection",
		}
	}

	return &UpdateCollectionResponse{
		Message:    "Collection updated successfully.",
		Collection: convert.CollectionModelToPayload(collection),
	}, nil
}

// DeleteCollectionParams is the parameters for deleting a collection
type DeleteCollectionParams struct {
	// The unique identifier for the collection
	ID uint64
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
	userData := auth.Data().(*UserData)

	collection, err := models.GetCollectionByID(ctx, params.ID, userData.ID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find collection",
		}
	} else if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find collection, unknown error",
		}
	}

	err = collection.Delete(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete collection",
		}
	}

	return &DeleteCollectionResponse{
		Message:    "Collection deleted successfully.",
		Collection: convert.CollectionModelToPayload(collection),
	}, nil
}
