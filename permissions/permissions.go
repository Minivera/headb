package permissions

import (
	"context"

	"encore.app/permissions/internal"
	"encore.app/permissions/models/generated/permissions/public/model"
)

// AddPermissionSetParams is the params to add a new permissions set to validate
// permissions on a key ID and optionally a database ID.
type AddPermissionSetParams struct {
	// The unique ID of the key to assign this permission to
	KeyID int64

	// The unique ID of the user assigned to this key and database. To validate
	// ownership when creating.
	UserID int64

	// The unique ID of the database to assign this permission to, if any
	DatabaseID *int64

	// The role to assign to this permission set
	Role string
}

// AddPermissionSetResponse is the response of the add permission set operation
type AddPermissionSetResponse struct {
	// The created permission set
	PermissionSet *model.Permissions
}

// AddPermissionSet adds a new permission set for an API key on an optional database.
// Will assign the role to the set and ignore any duplicates, since that means no
// permissions needs to be added.
//encore:api private
func AddPermissionSet(ctx context.Context, params *AddPermissionSetParams) (*AddPermissionSetResponse, error) {
	permissionSet, err := internal.AddPermissionSet(ctx, params.KeyID, params.UserID, params.DatabaseID, params.Role)
	if err != nil {
		return nil, err
	}

	return &AddPermissionSetResponse{
		PermissionSet: permissionSet,
	}, nil
}

// RemovePermissionSetParams is the params to remove a new permissions
type RemovePermissionSetParams struct {
	// The unique ID of the permission set to delete
	ID int64
}

// RemovePermissionSetResponse is the response of the remove permission set operation
type RemovePermissionSetResponse struct {
	// The deleted permission set
	PermissionSet *model.Permissions
}

// RemovePermissionSet removes a permission set using an ID.
//encore:api private
func RemovePermissionSet(ctx context.Context, params *RemovePermissionSetParams) (*RemovePermissionSetResponse, error) {
	permissionSet, err := internal.RemovePermissionSet(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &RemovePermissionSetResponse{
		PermissionSet: permissionSet,
	}, nil
}

// CanParams is the params to check if the key can act on the database (or on all databases if
// no ID is provided) given the operation.
type CanParams struct {
	// The unique ID of the key to assign this permission to
	KeyID int64

	// The unique ID of the database to assign this permission to, if any
	DatabaseID *int64

	// The operation to validate, should be a role
	Operation string
}

// CanResponse is the response of the can operation
type CanResponse struct {
	// Whether or not the key is allowed to do the operation
	Allowed bool
}

// Can validates if a key can take the provided operation on a database or all databases.
// FIXME: permissions are extremely atomic. Only a single permission can exists for a key ID and a
// database ID. Keys without database ID also do not count for permissions on the same key with
// a database ID.
//encore:api private
func Can(ctx context.Context, params *CanParams) (*CanResponse, error) {
	can, err := internal.Can(ctx, params.KeyID, params.DatabaseID, params.Operation)
	if err != nil {
		return nil, err
	}

	return &CanResponse{
		Allowed: can,
	}, nil
}
