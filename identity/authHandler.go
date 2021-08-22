package identity

import (
	"context"
	"strconv"

	"encore.app/identity/models"

	"encore.dev/beta/auth"
	log "github.com/sirupsen/logrus"
)

// UserData is the data containing the relevant information about the logged in user.
type UserData struct {
	KeyID    uint64
	ID       uint64
	Username string
	Token    string
}

// AuthHandler is our handler to fetch the user based on the passed api key.
//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *UserData, error) {
	response, err := GetUserForApiKeyInternal(ctx, &GetUserForApiKeyInternalParams{
		KeyString: token,
	})
	if err != nil || response == nil {
		log.WithError(err).Warning("Failed to authenticate user")
		return "", nil, nil
	}

	if response.User.Status == models.UserStatusPending {
		log.Warning("Authentication failed, user is still pending")
		return "", nil, nil
	}
	if response.User.Status == models.UserStatusDenied {
		log.Warning("Authentication failed, user is still pending")
		return "", nil, nil
	}

	return auth.UID(strconv.FormatUint(response.User.ID, 10)), &UserData{
		KeyID:    response.KeyID,
		ID:       response.User.ID,
		Token:    response.User.Token,
		Username: response.User.Username,
	}, nil
}
