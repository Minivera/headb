package identity

import (
	"context"
	"strconv"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encore.app/identity/keys"
	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
	"encore.app/identity/test_utils"
	test_utils2 "encore.app/test_utils"
)

func insertApiKey(ctx context.Context, apiKey *model.APIKeys) error {
	query, args := table.APIKeys.INSERT(
		table.APIKeys.ID,
		table.APIKeys.UserID,
		table.APIKeys.Value,
		table.APIKeys.LastUsedAt,
		table.APIKeys.UpdatedAt,
		table.APIKeys.CreatedAt,
	).VALUES(
		apiKey.ID,
		apiKey.UserID,
		apiKey.Value,
		apiKey.LastUsedAt,
		apiKey.UpdatedAt,
		apiKey.CreatedAt,
	).Sql()

	_, err := sqldb.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func TestGenerateApiKey(t *testing.T) {
	type expected struct {
		response *GenerateApiKeyResponse
		err      error
	}

	existingUser := &model.Users{
		ID:       1,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	tcs := []struct {
		scenario string
		userData *UserData
		expected expected
	}{
		{
			scenario: "Will generate a valid API Key for the user",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
			},
			expected: expected{
				response: &GenerateApiKeyResponse{},
			},
		},
		{
			scenario: "Will fail if the authenticated user is not found",
			userData: &UserData{
				ID: -1,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.Unauthenticated,
					Message: "Could not find user information, auth is invalid",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)

			defer test_utils.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			response, err := GenerateApiKey(ctx)
			if err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ApiKey)
			}
		})
	}
}

func TestListApiKeys(t *testing.T) {
	type expected struct {
		response *ListUserAPIKeysResponse
		err      error
	}

	existingUser := &model.Users{
		ID:       1,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	existingKey := &model.APIKeys{
		ID:         1,
		UserID:     existingUser.ID,
		Value:      "test",
		LastUsedAt: time.Now(),
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	tcs := []struct {
		scenario     string
		userData     *UserData
		existingKeys []*model.APIKeys
		expected     expected
	}{
		{
			scenario: "Will fetch the API keys for a user",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
			},
			existingKeys: []*model.APIKeys{existingKey},
			expected: expected{
				response: &ListUserAPIKeysResponse{
					Message: "Found 1 keys on this account.",
					Keys: []PublicKey{
						{
							KeyID:        existingKey.ID,
							LastUsedDate: existingKey.LastUsedAt,
							CreatedDate:  existingKey.CreatedAt,
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)

			defer test_utils.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			for _, k := range tc.existingKeys {
				err = insertApiKey(ctx, k)
				require.NoError(t, err)
			}

			response, err := ListApiKeys(ctx)
			if err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, response.Message, tc.expected.response.Message)

				assert.Len(t, response.Keys, len(tc.expected.response.Keys))
				for i, k := range response.Keys {
					assert.Equal(t, k.KeyID, tc.expected.response.Keys[i].KeyID)
				}
			}
		})
	}
}

func TestGetUserForApiKeyInternal(t *testing.T) {
	type expected struct {
		response *GetUserForApiKeyInternalResponse
		err      error
	}

	existingUser := &model.Users{
		ID:       1,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	keyValue, err := keys.GenerateApiKey()
	require.NoError(t, err)

	passwordKeyValue, err := keys.GenerateSecureApiKey(keyValue)
	require.NoError(t, err)

	now := time.Now()
	existingKey := &model.APIKeys{
		ID:         1,
		UserID:     existingUser.ID,
		Value:      passwordKeyValue,
		LastUsedAt: now,
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	keyString, err := keys.EncryptToPaseto(keyValue, existingKey.ID, secrets.SecretPasetoKey)
	require.NoError(t, err)

	notExistingKeyString, err := keys.EncryptToPaseto(keyValue, 1234, secrets.SecretPasetoKey)
	require.NoError(t, err)

	tcs := []struct {
		scenario string
		params   *GetUserForApiKeyInternalParams
		expected expected
	}{
		{
			scenario: "Will fetch the user for an API key",
			params: &GetUserForApiKeyInternalParams{
				KeyString: keyString,
			},
			expected: expected{
				response: &GetUserForApiKeyInternalResponse{
					KeyID: existingKey.ID,
					User:  existingUser,
				},
			},
		},
		{
			scenario: "Will fail if we could not parse the API key itself",
			params: &GetUserForApiKeyInternalParams{
				KeyString: "invalid%%%",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Could not parse API key",
				},
			},
		},
		{
			scenario: "Will fail if we could not find a key",
			params: &GetUserForApiKeyInternalParams{
				KeyString: notExistingKeyString,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.Internal,
					Message: "Could not find API key",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()

			defer test_utils.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			err = insertApiKey(ctx, existingKey)
			require.NoError(t, err)

			response, err := GetUserForApiKeyInternal(ctx, tc.params)
			if err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, response.KeyID, tc.expected.response.KeyID)
				assert.Equal(t, response.User.ID, tc.expected.response.User.ID)

				key, err := models.GetApiKey(ctx, response.KeyID)
				require.NoError(t, err)

				assert.NotEqual(t, key.LastUsedAt, now)
			}
		})
	}
}

func TestDeleteApiKey(t *testing.T) {
	type expected struct {
		response *DeleteApiKeyResponse
		err      error
	}

	existingUser := &model.Users{
		ID:       1,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	now := time.Now()
	existingKey := &model.APIKeys{
		ID:         1,
		UserID:     existingUser.ID,
		Value:      "test",
		LastUsedAt: now,
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	tcs := []struct {
		scenario string
		userData *UserData
		params   *DeleteApiKeyParams
		expected expected
	}{
		{
			scenario: "Will delete the API key by ID",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
			},
			params: &DeleteApiKeyParams{
				APIKeyID: existingKey.ID,
			},
			expected: expected{
				response: &DeleteApiKeyResponse{
					Message: "API key deleted, any calls made with it will not work anymore.",
				},
			},
		},
		{
			scenario: "Will fail if we could not find the API key",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
			},
			params: &DeleteApiKeyParams{
				APIKeyID: -1,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find API key",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)

			defer test_utils.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			err = insertApiKey(ctx, existingKey)
			require.NoError(t, err)

			response, err := DeleteApiKey(ctx, tc.params)
			log.Errorf("test %v", err)
			if err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, response.Message, tc.expected.response.Message)

				_, err := models.GetApiKey(ctx, tc.params.APIKeyID)
				assert.Error(t, err)
			}
		})
	}
}
