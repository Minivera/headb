package headb

import (
	"context"
	"errors"

	common2 "encore.app/headb/common"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"

	"encore.app/headb/models"
)

// GenerateApiKeyParams is the parameters for generating an API key
type GenerateApiKeyParams struct {
	// The username of the user to generate for
	Username string
}

// GenerateApiKeyResponse is the result of the generation of an API key
type GenerateApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The API key, this key cannot be obtained after being generated
	ApiKey string
}

// GenerateApiKey generates an API key for a specific user identified by it username. The user should
// have signed in before they generate a key.
//encore:api public
func GenerateApiKey(ctx context.Context, params *GenerateApiKeyParams) (*GenerateApiKeyResponse, error) {
	user, err := models.GetUserByUsername(ctx, params.Username)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find user for username, did you forget to run headb.SignIn?",
		}
	} else if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find user, unknown error",
		}
	}

	apiKey, err := common2.GenerateApiKey()
	if err != nil {
		log.Errorf("Could not generate API key, %v", err)
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not generate API key",
		}
	}

	hashedKey, err := common2.GenerateSecureApiKey(apiKey)
	if err != nil {
		log.Errorf("Could not generate API secure key to save in the database, %v", err)
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not generate API key",
		}
	}

	keyRecord := models.NewApiKey(hashedKey, user.ID)
	err = keyRecord.Save(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save API key",
		}
	}

	return &GenerateApiKeyResponse{
		Message: "Generated new API key for this user, we will not show this value again. Make sure to save it.",
		ApiKey:  common2.MergeWithKeyID(apiKey, keyRecord.ID),
	}, nil
}

// DeleteApiKeyResponse is the result of the deletion of an API key
type DeleteApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string
}

// DeleteApiKey will delete the API key used to authenticate this endpoint.
//encore:api auth
func DeleteApiKey(ctx context.Context) (*DeleteApiKeyResponse, error) {
	userData := auth.Data().(*UserData)

	apiKey, err := models.GetApiKey(ctx, userData.KeyID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find API key used to authenticate the request",
		}
	} else if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API key, unknown error",
		}
	}

	err = apiKey.Delete(ctx)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete API key",
		}
	}

	return &DeleteApiKeyResponse{
		Message: "API key deleted, any calls made with it will not work anymore.",
	}, nil
}
