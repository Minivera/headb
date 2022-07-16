package internal

import (
	"context"
	"fmt"

	"encore.app/content/helpers"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	log "github.com/sirupsen/logrus"

	"encore.app/content/convert"
	"encore.app/content/models"
	"encore.app/identity"
)

// StartTransaction creates a transaction on the provided database.
func StartTransaction(ctx context.Context, name string) (models.DatabasePayload, error) {
	userData := auth.Data().(*identity.UserData)

	if !helpers.CanAdmin(ctx, userData.KeyID) {
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	database := models.NewDatabase(name, userData.ID)
	if !models.ValidateDatabaseConstraint(ctx, database) {
		log.WithFields(map[string]interface{}{
			"name":    name,
			"user_id": userData.ID,
		}).Warning("Could not validate the constraints for the database, a database already exists.")
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A database with name `%s` already exists", database.Name),
		}
	}

	err := models.SaveDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not save databases")
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save database",
		}
	}

	return convert.DatabaseModelToPayload(database), nil
}

// UpdateDatabase updates a database by ID for the authenticated user
func UpdateDatabase(ctx context.Context, id uuid.UUID, name string) (convert.DatabasePayload, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, id, userData.ID)
	if err != nil {
		return convert.DatabasePayload{}, err
	}

	if !helpers.CanAdminDatabase(ctx, database.ID, userData.KeyID) {
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to administrate the database",
		}
	}

	database.Name = name
	if !models.ValidateDatabaseConstraint(ctx, database) {
		log.WithFields(map[string]interface{}{
			"name":    name,
			"user_id": userData.ID,
		}).Warning("Could not validate the constraints for the database, a database already exists.")
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A database with name `%s` already exists", database.Name),
		}
	}

	err = models.SaveDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not save database")
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save Database",
		}
	}

	return convert.DatabaseModelToPayload(database), nil
}

// DeleteDatabase deletes a database by ID for the authenticated user
func DeleteDatabase(ctx context.Context, id uuid.UUID) (convert.DatabasePayload, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, id, userData.ID)
	if err != nil {
		return convert.DatabasePayload{}, err
	}

	if !helpers.CanAdminDatabase(ctx, database.ID, userData.KeyID) {
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to administrate the database",
		}
	}

	err = models.DeleteDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not delete database")
		return convert.DatabasePayload{}, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete database",
		}
	}

	return convert.DatabaseModelToPayload(database), nil
}
