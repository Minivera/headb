package identity

import (
	"context"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	content_models "encore.app/content/models"
	test_utils_content "encore.app/content/test_utils"
	"encore.app/identity/keys"
	"encore.app/identity/models"
	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
	"encore.app/identity/test_utils"
	"encore.app/permissions"
	test_utils_permissions "encore.app/permissions/test_utils"
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingUser := &model.Users{
		ID:       newUUID,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	existingKey := &model.APIKeys{
		ID:         newUUID,
		UserID:     existingUser.ID,
		Value:      "test",
		LastUsedAt: time.Now(),
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	// Use models directly to avoid cyclic dependencies
	// FIXME: Fix this
	existingDatabase := content_models.NewDatabase("test", existingUser.ID)
	err = content_models.SaveDatabase(context.Background(), existingDatabase)
	require.NoError(t, err)

	defer test_utils_content.Cleanup(context.Background())

	tcs := []struct {
		scenario      string
		params        *GenerateApiKeyParams
		userData      *UserData
		usingAdminKey bool
		expected      expected
	}{
		{
			scenario: "Will generate a valid API Key for the user",
			params: &GenerateApiKeyParams{
				Role: "write",
			},
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			expected: expected{
				response: &GenerateApiKeyResponse{},
			},
		},
		{
			scenario: "Will generate a valid API Key for a database",
			params: &GenerateApiKeyParams{
				Role:       "write",
				DatabaseID: test_utils_permissions.UUIDPointer(existingDatabase.ID),
			},
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			expected: expected{
				response: &GenerateApiKeyResponse{},
			},
		},
		{
			scenario: "Will fail if not using an admin key",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: false,
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key cannot be used for admin operations",
				},
			},
		},
		{
			scenario: "Will fail if passing an invalid role",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			params: &GenerateApiKeyParams{
				Role: "invalid",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Role must be one of `write` or `read`",
				},
			},
		},
		{
			scenario: "Will fail if the authenticated user is not found",
			userData: &UserData{
				ID:    badUUID,
				KeyID: existingKey.ID,
			},
			usingAdminKey: true,
			params: &GenerateApiKeyParams{
				Role: "write",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.Unauthenticated,
					Message: "Could not find user information, auth is invalid",
				},
			},
		},
		{
			scenario: "Will fail if the database cannot be found",
			params: &GenerateApiKeyParams{
				Role:       "write",
				DatabaseID: test_utils_permissions.UUIDPointer(badUUID),
			},
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Database could not be found",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)

			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			err = insertApiKey(ctx, existingKey)
			require.NoError(t, err)

			if tc.usingAdminKey {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID: existingKey.ID,
					Role:  "admin",
				})
				require.NoError(t, err)
			}

			response, err := GenerateApiKey(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, response.ApiKey)

				keyResponse, err := GetUserForApiKeyInternal(ctx, &GetUserForApiKeyInternalParams{
					KeyString: response.ApiKey,
				})
				require.NoError(t, err)

				can, err := permissions.Can(ctx, &permissions.CanParams{
					KeyID:      keyResponse.KeyID,
					DatabaseID: tc.params.DatabaseID,
					Operation:  tc.params.Role,
				})
				require.NoError(t, err)

				assert.True(t, can.Allowed)
			}
		})
	}
}

func TestListApiKeys(t *testing.T) {
	type expected struct {
		response *ListUserAPIKeysResponse
		err      error
	}

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingUser := &model.Users{
		ID:       newUUID,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	existingKey := &model.APIKeys{
		ID:         newUUID,
		UserID:     existingUser.ID,
		Value:      "test",
		LastUsedAt: time.Now(),
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	tcs := []struct {
		scenario      string
		userData      *UserData
		usingAdminKey bool
		existingKeys  []*model.APIKeys
		expected      expected
	}{
		{
			scenario: "Will fetch the API keys for a user",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			existingKeys:  []*model.APIKeys{existingKey},
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
		{
			scenario: "Will fail if the user is not using an admin API key",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: false,
			existingKeys:  []*model.APIKeys{existingKey},
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key cannot be used for admin operations",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)

			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			for _, k := range tc.existingKeys {
				err = insertApiKey(ctx, k)
				require.NoError(t, err)

				if tc.usingAdminKey {
					_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
						KeyID: k.ID,
						Role:  "admin",
					})
					require.NoError(t, err)
				}
			}

			response, err := ListApiKeys(ctx)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingUser := &model.Users{
		ID:       newUUID,
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
		ID:         newUUID,
		UserID:     existingUser.ID,
		Value:      passwordKeyValue,
		LastUsedAt: now,
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	keyString, err := keys.EncryptToPaseto(keyValue, existingKey.ID, secrets.SecretPasetoKey)
	require.NoError(t, err)

	notExistingKeyString, err := keys.EncryptToPaseto(keyValue, badUUID, secrets.SecretPasetoKey)
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
			defer test_utils_permissions.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			err = insertApiKey(ctx, existingKey)
			require.NoError(t, err)

			response, err := GetUserForApiKeyInternal(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingUser := &model.Users{
		ID:       newUUID,
		Username: test_utils.StringPointer("test"),
		UniqueID: test_utils.StringPointer("1234"),
		Status:   model.UserStatus_Accepted,
	}

	now := time.Now()
	existingKey := &model.APIKeys{
		ID:         newUUID,
		UserID:     existingUser.ID,
		Value:      "test",
		LastUsedAt: now,
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	tcs := []struct {
		scenario      string
		userData      *UserData
		usingAdminKey bool
		params        *DeleteApiKeyParams
		expected      expected
	}{
		{
			scenario: "Will delete the API key by ID",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
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
			scenario: "Will fail if the key is not an admin key",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: false,
			params: &DeleteApiKeyParams{
				APIKeyID: existingKey.ID,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key cannot be used for admin operations",
				},
			},
		},
		{
			scenario: "Will fail if we could not find the API key",
			userData: &UserData{
				ID:       existingUser.ID,
				Username: *existingUser.Username,
				KeyID:    existingKey.ID,
			},
			usingAdminKey: true,
			params: &DeleteApiKeyParams{
				APIKeyID: badUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)

			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			err = insertApiKey(ctx, existingKey)
			require.NoError(t, err)

			if tc.usingAdminKey {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID: existingKey.ID,
					Role:  "admin",
				})
				require.NoError(t, err)
			}

			response, err := DeleteApiKey(ctx, tc.params)
			log.Errorf("test %v", err)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, response.Message, tc.expected.response.Message)

				_, err := models.GetApiKey(ctx, tc.params.APIKeyID)
				assert.Error(t, err)
			}
		})
	}
}
