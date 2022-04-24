package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/helpers"
	"encore.app/identity/keys"
	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/permissions"
	model_permissions "encore.app/permissions/models/generated/permissions/public/model"
)

// GenerateApiKeyParams are the params to generate a new API key with a role and, optionally, limited
// to a specific database.
type GenerateApiKeyParams struct {
	// The role to assign to this API Key, should be one of `write` or `read`
	Role string

	// An optional database ID to limit the api key to a specific database.
	DatabaseID *uuid.UUID
}

// GenerateApiKeyResponse is the result of the generation of an API key
type GenerateApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The API key, this key cannot be obtained after being generated
	ApiKey string
}

// GenerateApiKey generates a new API key for the authenticated user given the selected role. Passing an optional
// database ID allows the key to limit its operations to a specific database.
//encore:api auth
func GenerateApiKey(ctx context.Context, params *GenerateApiKeyParams) (*GenerateApiKeyResponse, error) {
	userData := auth.Data().(*UserData)

	can, err := permissions.Can(ctx, &permissions.CanParams{
		KeyID:     userData.KeyID,
		Operation: "admin",
	})
	if err != nil || !can.Allowed {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	if params.Role != model_permissions.Role_Write.String() && params.Role != model_permissions.Role_Read.String() {
		log.Errorf("Tried to create a new API key with role %s", params.Role)
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Role must be one of `write` or `read`",
		}
	}

	user, err := models.GetUserByID(ctx, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not fetch user from the auth ID")
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Could not find user information, auth is invalid",
		}
	}

	apiKey, key, err := createKeyForUser(ctx, user)
	if err != nil {
		log.WithError(err).Error("Could not create an API key for the user")
		return nil, err
	}

	_, err = permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
		UserID:     user.ID,
		KeyID:      key.ID,
		DatabaseID: params.DatabaseID,
		Role:       params.Role,
	})
	if err != nil {
		log.WithError(err).Error("Could not create permission set for api key")
		return nil, err
	}

	return &GenerateApiKeyResponse{
		Message: "Generated new API key, we will not show this key again. Make sure to save it.",
		ApiKey:  apiKey,
	}, nil
}

func createKeyForUser(ctx context.Context, user *model.Users) (string, *model.APIKeys, error) {
	apiKey, err := keys.GenerateApiKey()
	if err != nil {
		log.WithError(err).Error("Could not generate API key")
		return "", nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not generate API key",
		}
	}

	hashedKey, err := keys.GenerateSecureApiKey(apiKey)
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

	encryptedAPIKey, err := keys.EncryptToPaseto(apiKey, keyRecord.ID, secrets.SecretPasetoKey)
	if err != nil {
		log.WithError(err).Error("Could not encrypt the API key using paseto")
		return "", nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not process API key",
		}
	}

	return encryptedAPIKey, keyRecord, nil
}

// PublicKey is the representation of an API key that can safely be displayed
// through an external API request.
type PublicKey struct {
	// The unique identifier of the key, can be used to delete it.
	KeyID uuid.UUID

	// When this key was last used to authenticate a request.
	LastUsedDate time.Time

	// When this key was created.
	CreatedDate time.Time
}

// ListUserAPIKeysResponse is the result of fetching the all the API keys currently in
// effect for the authenticated user.
type ListUserAPIKeysResponse struct {
	// A message to inform the user of the result of the operation
	Message string

	// The keys found for the authenticated user.
	Keys []PublicKey
}

// ListApiKeys will list all the API keys available for the authenticated user.
//encore:api auth
func ListApiKeys(ctx context.Context) (*ListUserAPIKeysResponse, error) {
	userData := auth.Data().(*UserData)

	can, err := permissions.Can(ctx, &permissions.CanParams{
		KeyID:     userData.KeyID,
		Operation: "admin",
	})
	if err != nil || !can.Allowed {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	apiKeys, err := models.ListApiKeysForUser(ctx, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not fetch API keys for this user")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API keys",
		}
	}

	publicKeys := make([]PublicKey, len(apiKeys))
	for i, key := range apiKeys {
		publicKeys[i] = PublicKey{
			KeyID:        key.ID,
			LastUsedDate: key.LastUsedAt,
			CreatedDate:  key.CreatedAt,
		}
	}

	return &ListUserAPIKeysResponse{
		Message: fmt.Sprintf("Found %d keys on this account.", len(publicKeys)),
		Keys:    publicKeys,
	}, nil
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
	KeyID uuid.UUID

	// The fetched user identified for this API key.
	User *model.Users
}

// GetUserForApiKeyInternal finds the user for a given API key, given that it is valid and the
// user can access it.
//encore:api private
func GetUserForApiKeyInternal(ctx context.Context, params *GetUserForApiKeyInternalParams) (*GetUserForApiKeyInternalResponse, error) {
	keyValue, keyID, err := keys.ExtractIDAndValue(params.KeyString, secrets.SecretPasetoKey)
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
		log.WithError(err).Error("Could not find API key")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find API key",
		}
	}

	err = keys.ValidateKey(keyValue, apiKey.Value)
	if err != nil {
		log.WithError(err).Warning("Could not Validate the API key")
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Api key is invalid, user is not authorized",
		}
	}

	user, err := helpers.GetUser(ctx, apiKey.UserID)
	if err != nil {
		return nil, err
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

// DeleteApiKeyParams is the parameters for requesting the deletion of an API key for a user.
type DeleteApiKeyParams struct {
	// THe unique identifier of the key to delete
	APIKeyID uuid.UUID
}

// DeleteApiKeyResponse is the result of the deletion of an API key
type DeleteApiKeyResponse struct {
	// A message to inform the user of the result of the operation
	Message string
}

// DeleteApiKey will delete the API key used to authenticate this endpoint.
//encore:api auth
func DeleteApiKey(ctx context.Context, params *DeleteApiKeyParams) (*DeleteApiKeyResponse, error) {
	userData := auth.Data().(*UserData)

	can, err := permissions.Can(ctx, &permissions.CanParams{
		KeyID:     userData.KeyID,
		Operation: "admin",
	})
	if err != nil || !can.Allowed {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "API key cannot be used for admin operations",
		}
	}

	apiKey, err := helpers.GetApiKey(ctx, params.APIKeyID, userData.ID)
	if err != nil {
		log.WithError(err).Error("Could not find API key to delete")
		return nil, err
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
