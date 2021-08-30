package identity

import (
	"context"
	"testing"

	"encore.dev/beta/errs"
	"github.com/stretchr/testify/require"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"

	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
	"encore.app/identity/test_utils"
)

func insertUser(ctx context.Context, user *model.Users) error {
	query, args := table.Users.INSERT(
		table.Users.ID,
		table.Users.Username,
		table.Users.Token,
		table.Users.UniqueID,
		table.Users.Status,
		table.Users.UpdatedAt,
		table.Users.CreatedAt,
	).VALUES(
		user.ID,
		user.Username,
		user.Token,
		user.UniqueID,
		user.Status,
		user.UpdatedAt,
		user.CreatedAt,
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
			ctx := context.Background()

			defer test_utils.Cleanup(ctx)

			err := insertUser(ctx, existingUser)
			require.NoError(t, err)

			response, err := generateApiKey(ctx, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ApiKey)
			}
		})
	}
}
