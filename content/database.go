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

// ListDatabasesResponse is the list of databases for the current user
type ListDatabasesResponse struct {
	// The fetched databases
	Databases []convert.DatabasePayload
}

// ListDatabases lists all Databases created by the authenticated user.
//encore:api auth
func ListDatabases(ctx context.Context) (*ListDatabasesResponse, error) {
	userData := auth.Data().(*identity.UserData)

	if !helpers.CanAdmin(ctx, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	databases, err := models.ListDatabase(ctx, userData.ID)
	if err != nil {
		log.WithError(err).Warning("Could not fetch databases for the authenticated user")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not fetch databases",
		}
	}

	return &ListDatabasesResponse{
		Databases: convert.DatabaseModelsToPayloads(databases),
	}, nil
}

// GetDatabaseParams is the parameters for finding a database by ID
type GetDatabaseParams struct {
	// The unique identifier of the database
	ID int64
}

// GetDatabaseResponse is the result of having fetched a database
type GetDatabaseResponse struct {
	// The found database
	Database convert.DatabasePayload
}

// GetDatabase Finds a database by ID
//encore:api auth
func GetDatabase(ctx context.Context, params *GetDatabaseParams) (*GetDatabaseResponse, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanReadDatabase(ctx, database.ID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to read the database",
		}
	}

	return &GetDatabaseResponse{
		Database: convert.DatabaseModelToPayload(database),
	}, nil
}

// CreateDatabaseParams is the parameters for creating a database for collections
type CreateDatabaseParams struct {
	// The name of the database
	Name string
}

// CreateDatabaseResponse is the result of creating a database for collections
type CreateDatabaseResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The created database
	Database convert.DatabasePayload
}

// CreateDatabase creates a database for the authenticated user.
//encore:api auth
func CreateDatabase(ctx context.Context, params *CreateDatabaseParams) (*CreateDatabaseResponse, error) {
	userData := auth.Data().(*identity.UserData)

	if !helpers.CanAdmin(ctx, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	database := models.NewDatabase(params.Name, userData.ID)
	if !models.ValidateDatabaseConstraint(ctx, database) {
		log.WithFields(map[string]interface{}{
			"name":    params.Name,
			"user_id": userData.ID,
		}).Warning("Could not validate the constraints for the database, a database already exists.")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A database with name `%s` already exists", database.Name),
		}
	}

	err := models.SaveDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not save databases")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save database",
		}
	}

	return &CreateDatabaseResponse{
		Message:  "Database created successfully.",
		Database: convert.DatabaseModelToPayload(database),
	}, nil
}

// UpdateDatabaseParams is the parameters for updating a database
type UpdateDatabaseParams struct {
	// The unique identifier for the database
	ID int64

	// The name of the database
	Name string
}

// UpdateDatabaseResponse is the result of updating a database for collections
type UpdateDatabaseResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The updated database
	Database convert.DatabasePayload
}

// UpdateDatabase updates a database by ID for the authenticated user
//encore:api auth
func UpdateDatabase(ctx context.Context, params *UpdateDatabaseParams) (*UpdateDatabaseResponse, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanAdminDatabase(ctx, database.ID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to administrate the database",
		}
	}

	database.Name = params.Name
	if !models.ValidateDatabaseConstraint(ctx, database) {
		log.WithFields(map[string]interface{}{
			"name":    params.Name,
			"user_id": userData.ID,
		}).Warning("Could not validate the constraints for the database, a database already exists.")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("A database with name `%s` already exists", database.Name),
		}
	}

	err = models.SaveDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not save database")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save Database",
		}
	}

	return &UpdateDatabaseResponse{
		Message:  "Database updated successfully.",
		Database: convert.DatabaseModelToPayload(database),
	}, nil
}

// DeleteDatabaseParams is the parameters for deleting a database
type DeleteDatabaseParams struct {
	// The unique identifier for the database
	ID int64
}

// DeleteDatabaseResponse is the result of deleting a database for documents
type DeleteDatabaseResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The deleted database
	Database convert.DatabasePayload
}

// DeleteDatabase deletes a database by ID for the authenticated user
//encore:api auth
func DeleteDatabase(ctx context.Context, params *DeleteDatabaseParams) (*DeleteDatabaseResponse, error) {
	userData := auth.Data().(*identity.UserData)

	database, err := helpers.GetDatabase(ctx, params.ID, userData.ID)
	if err != nil {
		return nil, err
	}

	if !helpers.CanAdminDatabase(ctx, database.ID, userData.KeyID) {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key doesn't have the ability to administrate the database",
		}
	}

	err = models.DeleteDatabase(ctx, database)
	if err != nil {
		log.WithError(err).Error("Could not delete database")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete database",
		}
	}

	return &DeleteDatabaseResponse{
		Message:  "Database deleted successfully.",
		Database: convert.DatabaseModelToPayload(database),
	}, nil
}
