package content

import (
	"context"

	"encore.app/content/convert"
	"encore.app/content/internal"
)

// ListDatabasesResponse is the list of databases for the current user
type ListDatabasesResponse struct {
	// The fetched databases
	Databases []convert.DatabasePayload
}

// ListDatabases lists all Databases created by the authenticated user.
//encore:api auth
func ListDatabases(ctx context.Context) (*ListDatabasesResponse, error) {
	databases, err := internal.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}

	return &ListDatabasesResponse{
		Databases: databases,
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
	database, err := internal.GetDatabase(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &GetDatabaseResponse{
		Database: database,
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
	database, err := internal.CreateDatabase(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	return &CreateDatabaseResponse{
		Message:  "Database created successfully.",
		Database: database,
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
	database, err := internal.UpdateDatabase(ctx, params.ID, params.Name)
	if err != nil {
		return nil, err
	}

	return &UpdateDatabaseResponse{
		Message:  "Database updated successfully.",
		Database: database,
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
	database, err := internal.DeleteDatabase(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &DeleteDatabaseResponse{
		Message:  "Database deleted successfully.",
		Database: database,
	}, nil
}
