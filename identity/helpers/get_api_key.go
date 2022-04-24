package helpers

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
)

// GetApiKey gets an API key from a key ID and a user ID and returns a valid encore error
// if the key could not be fetched.
func GetApiKey(ctx context.Context, apiKeyID, userID uuid.UUID) (*model.APIKeys, error) {
	apiKey, err := models.GetApiKeyForUser(ctx, apiKeyID, userID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find an API key by the given ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find API key",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find API key")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API key, unknown error",
		}
	}

	return apiKey, nil
}
