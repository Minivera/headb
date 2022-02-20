package helpers

import (
	"context"

	log "github.com/sirupsen/logrus"

	"encore.app/permissions"
)

// CanAdmin checks if the given key ID can on as an admin all databases.
func CanAdmin(ctx context.Context, keyID int64) bool {
	can, err := permissions.Can(ctx, &permissions.CanParams{
		KeyID:     keyID,
		Operation: "admin",
	})
	if err == nil && can.Allowed {
		return true
	}

	log.WithError(err).Warning("Could not validate permissions for key without database ID, user is not allowed to act as admin")
	return false
}

// CanAdminDatabase checks if the given key ID can act as admin on the database, or on all database
// if no permissions can be validated for specific database.
func CanAdminDatabase(ctx context.Context, databaseID, keyID int64) bool {
	return canDoOnDatabase(ctx, "admin", databaseID, keyID)
}

// CanWriteDatabase checks if the given key ID can write on the database, or on all database
// if no permissions can be validated for specific database.
func CanWriteDatabase(ctx context.Context, databaseID, keyID int64) bool {
	return canDoOnDatabase(ctx, "write", databaseID, keyID)
}

// CanReadDatabase checks if the given key ID can read on the database, or on all database
// if no permissions can be validated for specific database.
func CanReadDatabase(ctx context.Context, databaseID, keyID int64) bool {
	return canDoOnDatabase(ctx, "read", databaseID, keyID)
}

func canDoOnDatabase(ctx context.Context, operation string, databaseID, keyID int64) bool {
	can, err := permissions.Can(ctx, &permissions.CanParams{
		KeyID:      keyID,
		DatabaseID: &databaseID,
		Operation:  operation,
	})

	if err == nil && can.Allowed {
		return true
	}

	log.WithFields(log.Fields{
		"database_id": databaseID,
		"operation":   operation,
	}).WithError(err).Warning("Could not validate permissions on database ID, trying without ID")

	can, err = permissions.Can(ctx, &permissions.CanParams{
		KeyID:     keyID,
		Operation: operation,
	})
	if err == nil && can.Allowed {
		return true
	}

	log.WithFields(log.Fields{
		"database_id": databaseID,
		"operation":   operation,
	}).WithError(err).Warningf("Could not validate permissions for key without database ID, user is not allowed to %s", operation)
	return false
}
