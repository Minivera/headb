package internal

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

// ListCollections lists all collections created by the authenticated user in the given
// database.
func ListCollections(ctx context.Context, databaseID int64) ([]convert.CollectionPayload, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, databaseID, userData.ID)
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

	return convert.CollectionModelsToPayloads(collections), nil
}

// GetCollection Finds a collection by ID
func GetCollection(ctx context.Context, id int64) (convert.CollectionPayload, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, id, userData.ID)
	if err != nil {
		return convert.CollectionPayload{}, err
	}

	if !helpers.CanReadDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	return convert.CollectionModelToPayload(collection), nil
}

// CreateCollection creates a collection for the given database if owned by the authenticated user.
func CreateCollection(ctx context.Context, databaseID int64, name string) (convert.CollectionPayload, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, databaseID, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not find database when creating a new collection")
		return convert.CollectionPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, database.ID, userData.KeyID) {
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	collection := models.NewCollection(name, database.ID)
	if !models.ValidateCollectionConstraint(ctx, collection) {
		log.WithFields(map[string]interface{}{
			"name":        name,
			"database_id": database.ID,
		}).Warning("Could not validate the constraints for collection, a collection already exists.")
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists in this database", collection.Name),
		}
	}

	err = models.SaveCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not save collections for this database")
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save collection",
		}
	}

	return convert.CollectionModelToPayload(collection), nil
}

// UpdateCollection updates a collection by ID for the authenticated user
func UpdateCollection(ctx context.Context, id int64, name string) (convert.CollectionPayload, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, id, userData.ID)
	if err != nil {
		return convert.CollectionPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	collection.Name = name
	if !models.ValidateCollectionConstraint(ctx, collection) {
		log.WithFields(map[string]interface{}{
			"name":        name,
			"database_id": collection.DatabaseID,
		}).Warning("Could not validate the constraints for collection, a collection already exists.")
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A collection with name `%s` already exists in this database", collection.Name),
		}
	}

	err = models.SaveCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not save collection")
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save collection",
		}
	}

	return convert.CollectionModelToPayload(collection), nil
}

// DeleteCollection deletes a collection by ID for the authenticated user
func DeleteCollection(ctx context.Context, id int64) (convert.CollectionPayload, error) {
	userData := auth.Data().(*identity.UserData)

	collection, err := helpers.GetCollection(ctx, id, userData.ID)
	if err != nil {
		return convert.CollectionPayload{}, err
	}

	if !helpers.CanWriteDatabase(ctx, collection.DatabaseID, userData.KeyID) {
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to write to the database",
		}
	}

	err = models.DeleteCollection(ctx, collection)
	if err != nil {
		log.WithError(err).Error("Could not delete collection")
		return convert.CollectionPayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete collection",
		}
	}

	return convert.CollectionModelToPayload(collection), nil
}
