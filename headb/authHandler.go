package headb

import (
	"context"
	"errors"
	"strconv"

	common2 "encore.app/headb/common"

	"encore.app/headb/models"
	"encore.dev/storage/sqldb"

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
	keyValue, keyID, err := common2.ExtractIDAndValue(token)
	if err != nil {
		log.Errorf("Tried to extract API key value and ID from token, %v", err)
		return "", nil, nil
	}

	apiKey, err := models.GetApiKey(ctx, keyID)
	if errors.Is(err, sqldb.ErrNoRows) {
		log.Errorf("Could not find API key record from token, %v", err)
		return "", nil, nil
	} else if err != nil {
		return "", nil, err
	}

	err = common2.ValidateKey(keyValue, apiKey.Value)
	if err != nil {
		log.Errorf("API key value did not match found API key, %v", err)
		return "", nil, nil
	}

	user, err := models.GetUserByID(ctx, apiKey.UserID)
	if errors.Is(err, sqldb.ErrNoRows) {
		log.Errorf("Coult not find user from associated to API key, %v", err)
		return "", nil, nil
	} else if err != nil {
		return "", nil, err
	}

	return auth.UID(strconv.FormatUint(user.ID, 10)), &UserData{
		KeyID:    keyID,
		ID:       user.ID,
		Token:    user.Token,
		Username: user.Username,
	}, nil
}
