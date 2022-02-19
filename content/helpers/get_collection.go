package helpers

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models"
	"encore.app/content/models/generated/content/public/model"
)

// GetCollection gets a collection from a collection ID and a user ID and returns a valid encore error
// if the collection could not be fetched.
func GetCollection(ctx context.Context, collectionID, userID int64) (*model.Collections, error) {
	collection, err := models.GetCollectionByID(ctx, collectionID, userID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find collection by ID for document")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find collection",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find collection for document")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find collection, unknown error",
		}
	}

	return collection, nil
}
