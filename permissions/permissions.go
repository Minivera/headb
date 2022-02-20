package permissions

import (
	"context"
	"errors"

	"encore.app/permissions/models"
	"encore.app/permissions/models/generated/permissions/public/model"
	"encore.dev/beta/errs"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"
)

// AddPermissionSetParams is the params to add a new permissions set to validate
// permissions on a key ID and optionally a database ID.
type AddPermissionSetParams struct {
	// The unique ID of the key to assign this permission to
	KeyID int64

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
	role := model.Role("")
	err := role.Scan(params.Role)
	if err != nil {
		log.WithError(err).Warning("Selected role is not valid")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
		}
	}

	permissionSet := models.NewPermissionSet(params.KeyID, params.DatabaseID, role)
	err = models.CreatePermissionSet(ctx, permissionSet)
	if err != nil {
		log.WithFields(log.Fields{
			"key_id":      params.KeyID,
			"database_id": params.DatabaseID,
			"role":        role,
		}).WithError(err).Error("Could not save permission set")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "Could not save permission set, set probably already exists",
		}
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
	permissionSet, err := models.GetPermissionByID(ctx, params.ID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find permission by the given ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find permission set",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find permission by the given ID")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find permission set",
		}
	}

	err = models.DeletePermissionSet(ctx, permissionSet)
	if err != nil {
		log.WithError(err).Error("Could not delete permission set")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete permission set",
		}
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
	operation := model.Role("")
	err := operation.Scan(params.Operation)
	if err != nil {
		log.WithError(err).Warning("Given operation is not valid")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
		}
	}

	permissionSet, err := models.GetPermission(ctx, params.KeyID, params.DatabaseID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find permission, returning unallowed")
		return &CanResponse{
			Allowed: false,
		}, nil
	} else if err != nil {
		log.WithError(err).Error("Could not find permission")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find permission set",
		}
	}

	can := false
	switch operation {
	case model.Role_Admin:
		can = permissionSet.Role == model.Role_Admin
	case model.Role_Write:
		can = permissionSet.Role == model.Role_Admin || permissionSet.Role == model.Role_Write
	case model.Role_Read:
		can = permissionSet.Role == model.Role_Admin ||
			permissionSet.Role == model.Role_Write ||
			permissionSet.Role == model.Role_Read
	}

	return &CanResponse{
		Allowed: can,
	}, nil
}
