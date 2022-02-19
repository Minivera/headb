package helpers

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
)

// GetUser gets a user from a user ID and returns a valid encore error
// if the user could not be fetched.
func GetUser(ctx context.Context, userID int64) (*model.Users, error) {
	user, err := models.GetUserByID(ctx, userID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find user by the given ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find user for API key",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find user")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find user for API key",
		}
	}

	return user, nil
}
