package content

import (
	"context"
	"fmt"

	"encore.app/content/helpers"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	log "github.com/sirupsen/logrus"

	"encore.app/content/convert"
	"encore.app/content/models"
	"encore.app/identity"
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
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, params.DatabaseID, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not find database when listing collections")
		return nil, err
	}

	if !helpers.CanReadDatabase(ctx, database.ID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	collections, err := models.ListCollections(ctx, database.ID)
	if err != nil {
		log.WithError(err).Warning("Could not fetch collections for this database")
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
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanReadDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	return &GetCollectionResponse{
		Collection: convert.CollectionModelToPayload(collection),
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
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, params.DatabaseID, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not find database when creating a new collection")
		return nil, err
	}

	if !helpers.CanWriteDatabase(ctx, database.ID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	collection := models.NewCollection(params.Name, database.ID)
	if !models.ValidateCollectionConstraint(ctx, collection) {
		log.WithFields(map[string]interface{}{
			"name":        params.Name,
			"database_id": database.ID,
		}).Warning("Could not validate the constraints for collection, a collection already exists.")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists in this database", collection.Name),
		}
	}

	err = models.SaveCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not save collections for this database")
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
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	collection.Name = params.Name
	if !models.ValidateCollectionConstraint(ctx, collection) {
		log.WithFields(map[string]interface{}{
			"name":        params.Name,
			"database_id": collection.DatabaseID,
		}).Warning("Could not validate the constraints for collection, a collection already exists.")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists in this database", collection.Name),
		}
	}

	err = models.SaveCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not save collection")
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
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	err = models.DeleteCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not delete collection")
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
