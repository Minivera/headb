package permissions

import (
	"context"
	"testing"
	"time"

	"encore.app/permissions/models"
	"encore.app/permissions/models/generated/permissions/public/model"
	"encore.app/permissions/models/generated/permissions/public/table"
	"encore.app/permissions/test_utils"
	test_utils2 "encore.app/test_utils"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertPermissions(ctx context.Context, permissions []*model.Permissions) error {
	for _, permission := range permissions {
		query, args := table.Permissions.INSERT(
			table.Permissions.ID,
			table.Permissions.KeyID,
			table.Permissions.DatabaseID,
			table.Permissions.Role,
			table.Permissions.UpdatedAt,
			table.Permissions.CreatedAt,
		).VALUES(
			permission.ID,
			permission.KeyID,
			permission.DatabaseID,
			permission.Role,
			permission.UpdatedAt,
			permission.CreatedAt,
		).Sql()

		_, err := sqldb.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAddPermissionSet(t *testing.T) {
	type expected struct {
		response *AddPermissionSetResponse
		err      error
	}

	existingPermissions := []*model.Permissions{
		{
			KeyID:      1,
			DatabaseID: test_utils.Int64Pointer(1),
			Role:       "admin",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	tcs := []struct {
		scenario string
		params   *AddPermissionSetParams
		expected expected
	}{
		{
			scenario: "Will create a permissions set with a database ID",
			params: &AddPermissionSetParams{
				KeyID:      1234,
				DatabaseID: test_utils.Int64Pointer(1234),
				Role:       "read",
			},
			expected: expected{
				response: &AddPermissionSetResponse{
					PermissionSet: &model.Permissions{
						KeyID:      1234,
						DatabaseID: test_utils.Int64Pointer(1234),
						Role:       "read",
					},
				},
			},
		},
		{
			scenario: "Will create a permissions set without a database ID",
			params: &AddPermissionSetParams{
				KeyID: 1234,
				Role:  "read",
			},
			expected: expected{
				response: &AddPermissionSetResponse{
					PermissionSet: &model.Permissions{
						KeyID: 1234,
						Role:  "read",
					},
				},
			},
		},
		{
			scenario: "Will fail when the role is invalid",
			params: &AddPermissionSetParams{
				KeyID: 1234,
				Role:  "invalid",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
				},
			},
		},
		{
			scenario: "Will fail if the set already exists",
			params: &AddPermissionSetParams{
				KeyID:      1,
				DatabaseID: test_utils.Int64Pointer(1),
				Role:       "read",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "Could not save permission set, set probably already exists",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertPermissions(ctx, existingPermissions)
			require.NoError(t, err)

			response, err := AddPermissionSet(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.PermissionSet.KeyID, response.PermissionSet.KeyID)
				if tc.expected.response.PermissionSet.DatabaseID == nil {
					require.Nil(t, response.PermissionSet.DatabaseID)
				} else {
					require.NotNil(t, response.PermissionSet.DatabaseID)
					assert.Equal(t, *tc.expected.response.PermissionSet.DatabaseID, *response.PermissionSet.DatabaseID)
				}
				assert.Equal(t, tc.expected.response.PermissionSet.Role, response.PermissionSet.Role)
			}
		})
	}
}

func TestRemovePermissionSet(t *testing.T) {
	type expected struct {
		response *RemovePermissionSetResponse
		err      error
	}

	existingPermissions := []*model.Permissions{
		{
			ID:         1,
			KeyID:      1,
			DatabaseID: test_utils.Int64Pointer(1),
			Role:       "admin",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	tcs := []struct {
		scenario string
		params   *RemovePermissionSetParams
		expected expected
	}{
		{
			scenario: "Will remove a permissions set from an ID",
			params: &RemovePermissionSetParams{
				ID: existingPermissions[0].ID,
			},
			expected: expected{
				response: &RemovePermissionSetResponse{
					PermissionSet: existingPermissions[0],
				},
			},
		},
		{
			scenario: "Will fail if the set cannot be found",
			params: &RemovePermissionSetParams{
				ID: -1,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find permission set",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertPermissions(ctx, existingPermissions)
			require.NoError(t, err)

			response, err := RemovePermissionSet(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.PermissionSet.ID, response.PermissionSet.ID)

				_, err := models.GetPermissionByID(ctx, tc.params.ID)
				require.Error(t, err)
			}
		})
	}
}

func TestCan(t *testing.T) {
	type expected struct {
		response *CanResponse
		err      error
	}

	now := time.Now()
	tcs := []struct {
		scenario            string
		params              *CanParams
		existingPermissions []*model.Permissions
		expected            expected
	}{
		{
			scenario: "Will return true when checking if an admin can admin",
			params: &CanParams{
				KeyID:     1,
				Operation: "admin",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "admin",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return true when checking if an admin can write",
			params: &CanParams{
				KeyID:     1,
				Operation: "write",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "admin",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return true when checking if an admin can read",
			params: &CanParams{
				KeyID:     1,
				Operation: "read",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "admin",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return false when checking if a writer can admin",
			params: &CanParams{
				KeyID:     1,
				Operation: "admin",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "write",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: false,
				},
			},
		},
		{
			scenario: "Will return true when checking if a writer can write",
			params: &CanParams{
				KeyID:     1,
				Operation: "write",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "write",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return true when checking if a writer can read",
			params: &CanParams{
				KeyID:     1,
				Operation: "read",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "write",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return false when checking if a reader can admin",
			params: &CanParams{
				KeyID:     1,
				Operation: "admin",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "read",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: false,
				},
			},
		},
		{
			scenario: "Will return false when checking if a reader can write",
			params: &CanParams{
				KeyID:     1,
				Operation: "write",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "read",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: false,
				},
			},
		},
		{
			scenario: "Will return true when checking if a reader can read",
			params: &CanParams{
				KeyID:     1,
				Operation: "read",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:        1,
					KeyID:     1,
					Role:      "read",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return true when checking for an Allowed operation for a key on a database ID",
			params: &CanParams{
				KeyID:      1,
				DatabaseID: test_utils.Int64Pointer(1),
				Operation:  "read",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:         1,
					KeyID:      1,
					DatabaseID: test_utils.Int64Pointer(1),
					Role:       "write",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: true,
				},
			},
		},
		{
			scenario: "Will return false when checking an Allowed operation for a key on the wrong database ID",
			params: &CanParams{
				KeyID:      1,
				DatabaseID: test_utils.Int64Pointer(2),
				Operation:  "read",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:         1,
					KeyID:      1,
					DatabaseID: test_utils.Int64Pointer(1),
					Role:       "write",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: false,
				},
			},
		},
		{
			scenario: "Will return false when checking an unAllowed operation for a key on a database ID",
			params: &CanParams{
				KeyID:      1,
				DatabaseID: test_utils.Int64Pointer(1),
				Operation:  "admin",
			},
			existingPermissions: []*model.Permissions{
				{
					ID:         1,
					KeyID:      1,
					DatabaseID: test_utils.Int64Pointer(1),
					Role:       "write",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			expected: expected{
				response: &CanResponse{
					Allowed: false,
				},
			},
		},
		{
			scenario: "Will fail if given an invalid operation",
			params: &CanParams{
				KeyID:     1,
				Operation: "invalid",
			},
			existingPermissions: []*model.Permissions{},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Selected role is not valid, must be one of `admin`, `write`, or `read`",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertPermissions(ctx, tc.existingPermissions)
			require.NoError(t, err)

			response, err := Can(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Allowed, response.Allowed)
			}
		})
	}
}
