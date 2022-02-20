package identity

import (
	"context"
	"strconv"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models/generated/identity/public/model"
)

// UserData is the data containing the relevant information about the logged in user.
type UserData struct {
	KeyID    int64
	ID       int64
	Username string
	Token    string
}

// AuthHandler is our handler to fetch the user based on the passed api key.
//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *UserData, error) {
	if token == "" {
		return "", nil, nil
	}

	response, err := GetUserForApiKeyInternal(ctx, &GetUserForApiKeyInternalParams{
		KeyString: token,
	})
	if err != nil || response == nil {
		log.WithError(err).Warning("Failed to authenticate user")
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Could not authenticate with the given API key",
		}
	}

	if response.User.Status == model.UserStatus_Pending {
		log.Warning("Authentication failed, user is still pending")
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
			Message: "User trying to authenticate is still pending, " +
				"make sure to complete the sign-in process before using the API key",
		}
	}
	if response.User.Status == model.UserStatus_Denied {
		log.Warning("Authentication failed, user is denied")
		return "", nil, &errs.Error{
			Code: errs.Unauthenticated,
			Message: "User trying to authenticate was denied, " +
				"please try to sign-in again before using the API key",
		}
	}

	return auth.UID(strconv.FormatInt(response.User.ID, 10)), &UserData{
		KeyID:    response.KeyID,
		ID:       response.User.ID,
		Token:    *response.User.Token,
		Username: *response.User.Username,
	}, nil
}
