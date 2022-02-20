package models

import (
	"context"

	"encore.dev/storage/sqldb"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/permissions/models/generated/permissions/public/model"
	"encore.app/permissions/models/generated/permissions/public/table"
)

var db = sqldb.Named("permissions").Stdlib()

// NewPermissionSet generates a new PermissionSet structure using the given unique IDs.
func NewPermissionSet(keyID int64, databaseID *int64, role model.Role) *model.Permissions {
	return &model.Permissions{
		KeyID:      keyID,
		DatabaseID: databaseID,
		Role:       role,
	}
}

// GetPermissionByID fetches a single permissionSet by ID. Returns nil on an error.
func GetPermissionByID(ctx context.Context, id int64) (*model.Permissions, error) {
	statement := postgres.SELECT(
		table.Permissions.ID,
		table.Permissions.KeyID,
		table.Permissions.DatabaseID,
		table.Permissions.Role,
		table.Permissions.UpdatedAt,
		table.Permissions.CreatedAt,
	).FROM(
		table.Permissions,
	).WHERE(
		table.Permissions.ID.EQ(postgres.Int64(id)),
	).LIMIT(1)

	permissionSet := model.Permissions{}
	err := statement.QueryContext(ctx, db, &permissionSet)
	if err != nil {
		log.WithError(err).Error("Could not query permissionSet")
		return nil, err
	}

	return &permissionSet, nil
}

// GetPermission fetches a single permissionSet record given a key ID and an optional
// database ID of the database may be assigned to. Returns nil on an error.
func GetPermission(ctx context.Context, keyID int64, databaseID *int64) (*model.Permissions, error) {
	condition := table.Permissions.KeyID.EQ(postgres.Int64(keyID))
	if databaseID != nil {
		condition = condition.AND(table.Permissions.DatabaseID.EQ(postgres.Int64(*databaseID)))
	}

	statement := postgres.SELECT(
		table.Permissions.ID,
		table.Permissions.KeyID,
		table.Permissions.DatabaseID,
		table.Permissions.Role,
		table.Permissions.UpdatedAt,
		table.Permissions.CreatedAt,
	).FROM(
		table.Permissions,
	).WHERE(
		condition,
	).LIMIT(1)

	permissionSet := model.Permissions{}
	err := statement.QueryContext(ctx, db, &permissionSet)
	if err != nil {
		log.WithError(err).Error("Could not query permissionSet")
		return nil, err
	}

	return &permissionSet, nil
}

// CreatePermissionSet create a permission set it is called with in the database.
// Will throw an error if constraints fail or the permission cannot be inserted.
func CreatePermissionSet(ctx context.Context, permissionSet *model.Permissions) error {
	statement := table.Permissions.INSERT(
		table.Permissions.KeyID,
		table.Permissions.Role,
		table.Permissions.DatabaseID,
	)

	if permissionSet.DatabaseID != nil {
		statement.VALUES(
			postgres.Int64(permissionSet.KeyID),
			permissionSet.Role,
			postgres.Int64(*permissionSet.DatabaseID),
		)
	} else {
		statement.VALUES(
			postgres.Int64(permissionSet.KeyID),
			permissionSet.Role,
			nil,
		)
	}

	query, args := statement.RETURNING(
		table.Permissions.ID,
		table.Permissions.UpdatedAt,
		table.Permissions.CreatedAt,
	).Sql()

	err := sqldb.
		QueryRow(ctx, query, args...).
		Scan(&permissionSet.ID, &permissionSet.UpdatedAt, &permissionSet.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not insert permissionSet")
		return err
	}

	return nil

}

// DeletePermissionSet deletes the PermissionSet is it called on.
func DeletePermissionSet(ctx context.Context, permissionSet *model.Permissions) error {
	query, args := table.Permissions.
		DELETE().
		WHERE(table.Permissions.ID.EQ(postgres.Int64(permissionSet.ID))).
		RETURNING(table.Permissions.ID).
		Sql()

	deletedID := 0
	err := sqldb.QueryRow(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == 0 {
		log.WithError(err).Error("Could not delete permissionSet")
		return err
	}

	return nil
}
