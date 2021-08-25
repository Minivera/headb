package identity

import (
	"context"
	"errors"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/helpers"
	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
)

// GenerateApiKeyResponse is the result of the generation of an API key
type GenerateApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The API key, this key cannot be obtained after being generated
	ApiKey string
}

// GenerateApiKey generates an API key for a specific user identified by it username. The user should
// have signed in before they generate a key.
//encore:api auth
func GenerateApiKey(ctx context.Context) (*GenerateApiKeyResponse, error) {
	userData := auth.Data().(*UserData)

	user, err := models.GetUserByID(ctx, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not fetch user from the auth ID")
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Could not find user information, auth is invalid",
		}
	}

	apiKey, keyRecord, err := createKeyForUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not create an API key for the user")
		return nil, err
	}

	// TODO: Merging the key with the ID is not a great idea, we should try to use a better
	// method to authenticate users with an API key. Maybe JWTs?
	return &GenerateApiKeyResponse{
		Message: "Generated new API key, we will not show this key again. Make sure to save it.",
		ApiKey:  helpers.MergeWithKeyID(apiKey, keyRecord.ID),
	}, nil
}

func createKeyForUser(ctx context.Context, user *model.Users) (string, *model.APIKeys, error) {
	apiKey, err := helpers.GenerateApiKey()
	if err != nil {
		log.WithError(err).Error("Could not generate API key")
		return "", nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not generate API key",
		}
	}

	hashedKey, err := helpers.GenerateSecureApiKey(apiKey)
	if err != nil {
		log.WithError(err).Error("Could not generate API secure key to save in the database")
		return "", nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not generate API key",
		}
	}

	keyRecord := models.NewApiKey(hashedKey, user.ID)
	err = models.SaveApiKey(ctx, keyRecord)
	if err != nil {
		log.WithError(err).Error("Could not save the generated API key")
		return "", nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not save API key",
		}
	}

	return apiKey, keyRecord, nil
}

// GetUserForApiKeyInternalParams is the parameters for fetching an user for authentication between
// services using an API key.
type GetUserForApiKeyInternalParams struct {
	// The API key used for authentication
	KeyString string
}

// GetUserForApiKeyInternalResponse is the result of fetching the user using an API key.
type GetUserForApiKeyInternalResponse struct {
	// The unique identifier of the key given in the request, for later use.
	KeyID int64

	// The fetched user identified for this API key.
	User *model.Users
}

// GetUserForApiKeyInternal finds the user for a given API key, given that it is valid and the
// user can access it.
//encore:api private
func GetUserForApiKeyInternal(ctx context.Context, params *GetUserForApiKeyInternalParams) (*GetUserForApiKeyInternalResponse, error) {
	keyValue, keyID, err := helpers.ExtractIDAndValue(params.KeyString)
	if err != nil {
		log.WithError(err).Warning("Could not parse API key, error when extracting the ID and value")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Could not parse API key",
		}
	}

	apiKey, err := models.GetApiKey(ctx, keyID)
	if errors.Is(err, sqldb.ErrNoRows) {
		log.WithError(err).Warning("Could not find an API key for the given ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find API key",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find APi key")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API key",
		}
	}

	err = helpers.ValidateKey(keyValue, apiKey.Value)
	if err != nil {
		log.WithError(err).Warning("Could not Validate the API key")
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Api key is invalid, user is not authorized",
		}
	}

	user, err := models.GetUserByID(ctx, apiKey.UserID)
	if errors.Is(err, sqldb.ErrNoRows) {
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

	// Save the key in order to update the last_used_at date
	// Ignore any potential error, but do log them. We don't care about the date being unsaved
	err = models.SaveApiKey(ctx, apiKey)
	if err != nil {
		log.WithError(err).Warning("Could not save the API key")
	}

	return &GetUserForApiKeyInternalResponse{
		KeyID: keyID,
		User:  user,
	}, nil
}

// DeleteApiKeyResponse is the result of the deletion of an API key
type DeleteApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string
}

// DeleteApiKey will delete the API key used to authenticate this endpoint.
// TODO: Add some endpoint to list API keys and delete them
//encore:api auth
func DeleteApiKey(ctx context.Context) (*DeleteApiKeyResponse, error) {
	userData := auth.Data().(*UserData)

	// TODO: Delete an API key by ID, not the current API key. That
	// could lock a user into having no API key to create an API key with.
	apiKey, err := models.GetApiKey(ctx, userData.KeyID)
	if errors.Is(err, sqldb.ErrNoRows) {
		log.WithError(err).Warning("Could not find an API key by the given ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find API key used to authenticate the request",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find API key")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API key, unknown error",
		}
	}

	err = models.DeleteApiKey(ctx, apiKey)
	if err != nil {
		log.WithError(err).Error("Could not delete the API key")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not delete API key",
		}
	}

	return &DeleteApiKeyResponse{
		Message: "API key deleted, any calls made with it will not work anymore.",
	}, nil
}
