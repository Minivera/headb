package helpers

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models"
	"encore.app/content/models/generated/content/public/model"
)

// GetDatabase gets a database from a database ID and a user ID and returns a valid encore error
// if the database could not be fetched.
func GetDatabase(ctx context.Context, databaseID uuid.UUID, userID uuid.UUID) (*model.Databases, error) {
	database, err := models.GetDatabaseByID(ctx, databaseID, userID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find database by ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find database",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find database")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find database, unknown error",
		}
	}

	return database, nil
}
