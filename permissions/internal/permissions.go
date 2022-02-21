package internal

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	content_models "encore.app/content/models"
	"encore.app/permissions/models"
	"encore.app/permissions/models/generated/permissions/public/model"
)

// AddPermissionSet adds a new permission set for an API key on an optional database.
// Will assign the role to the set and ignore any duplicates, since that means no
// permissions needs to be added.
func AddPermissionSet(ctx context.Context, keyID, userID int64, databaseID *int64, givenRole string) (*model.Permissions, error) {
	role := model.Role("")
	err := role.Scan(givenRole)
	if err != nil {
		log.WithError(err).Warning("Selected role is not valid")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
		}
	}

	if databaseID != nil {
		// Get the database if the key was created for a specific database
		// use the models directly here to avoid cycling dependencies.
		_, err := content_models.GetDatabaseByID(ctx, *databaseID, userID)
		if err != nil {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Database could not be found",
			}
		}
	}

	permissionSet := models.NewPermissionSet(keyID, databaseID, role)
	err = models.CreatePermissionSet(ctx, permissionSet)
	if err != nil {
		log.WithFields(log.Fields{
			"key_id":      keyID,
			"database_id": databaseID,
			"role":        role,
		}).WithError(err).Error("Could not save permission set")
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "Could not save permission set, set probably already exists",
		}
	}

	return permissionSet, nil
}

// RemovePermissionSet removes a permission set using an ID.
func RemovePermissionSet(ctx context.Context, id int64) (*model.Permissions, error) {
	permissionSet, err := models.GetPermissionByID(ctx, id)
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

	return permissionSet, nil
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
func Can(ctx context.Context, keyID int64, databaseID *int64, givenOperation string) (bool, error) {
	operation := model.Role("")
	err := operation.Scan(givenOperation)
	if err != nil {
		log.WithError(err).Warning("Given operation is not valid")
		return false, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
		}
	}

	permissionSet, err := models.GetPermission(ctx, keyID, databaseID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find permission, returning unallowed")
		return false, nil
	} else if err != nil {
		log.WithError(err).Error("Could not find permission")
		return false, &errs.Error{
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

	return can, nil
}
