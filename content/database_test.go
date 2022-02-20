package content

import (
	"context"
	"strconv"
	"testing"
	"time"

	test_utils2 "encore.app/test_utils"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"encore.app/content/convert"
	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
	"encore.app/content/test_utils"
	"encore.app/identity"
)

func compareDatabases(t *testing.T, expected, actual []convert.DatabasePayload) {
	assert.Len(t, actual, len(expected))
	for _, database := range expected {
		var found *convert.DatabasePayload
		for _, lookup := range actual {
			if lookup.ID == database.ID {
				found = &lookup
				break
			}
		}

		require.NotNilf(t, found, "Could not find database with ID %d in actual", database.ID)

		assert.Equal(t, database.Name, found.Name)
	}
}

func insertDatabases(ctx context.Context, databases []*model.Databases) error {
	for _, database := range databases {
		query, args := table.Databases.INSERT(
			table.Databases.ID,
			table.Databases.Name,
			table.Databases.UserID,
			table.Databases.UpdatedAt,
			table.Databases.CreatedAt,
		).VALUES(
			database.ID,
			database.Name,
			database.UserID,
			database.UpdatedAt,
			database.CreatedAt,
		).Sql()

		_, err := sqldb.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestListDatabases(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *ListDatabasesResponse
		err      error
	}

	validDatabases := []*model.Databases{
		{
			ID:        2,
			UserID:    1,
			Name:      "test",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        3,
			UserID:    1,
			Name:      "test2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDatabases []*model.Databases
		expected          expected
	}{
		{
			scenario:          "Returns a list of databases owned by a user",
			userData:          &identity.UserData{ID: 1},
			existingDatabases: validDatabases,
			expected: expected{
				response: &ListDatabasesResponse{
					Databases: convert.DatabaseModelsToPayloads(validDatabases),
				},
			},
		},
		{
			scenario:          "Returns empty when the user owns no databases",
			userData:          &identity.UserData{ID: 2},
			existingDatabases: validDatabases,
			expected: expected{
				response: &ListDatabasesResponse{
					Databases: []convert.DatabasePayload{},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)

			err := insertDatabases(ctx, tc.existingDatabases)
			require.NoError(t, err)

			response, err := ListDatabases(ctx)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				compareDatabases(t, tc.expected.response.Databases, response.Databases)
			}
		})
	}
}

func TestGetDatabase(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *GetDatabaseResponse
		err      error
	}

	validDatabases := []*model.Databases{
		{
			ID:        2,
			UserID:    1,
			Name:      "test",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		params            *GetDatabaseParams
		existingDatabases []*model.Databases
		expected          expected
	}{
		{
			scenario:          "Returns a database by ID, owned by a user",
			userData:          &identity.UserData{ID: 1},
			params:            &GetDatabaseParams{ID: validDatabases[0].ID},
			existingDatabases: validDatabases,
			expected: expected{
				response: &GetDatabaseResponse{
					Database: convert.DatabaseModelToPayload(validDatabases[0]),
				},
			},
		},
		{
			scenario:          "Returns an error when the database is not found",
			userData:          &identity.UserData{ID: 1},
			params:            &GetDatabaseParams{ID: 3},
			existingDatabases: validDatabases,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
		{
			scenario:          "Returns an error when the user does not own the database",
			userData:          &identity.UserData{ID: 2},
			params:            &GetDatabaseParams{ID: validDatabases[0].ID},
			existingDatabases: validDatabases,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)

			err := insertDatabases(ctx, tc.existingDatabases)
			require.NoError(t, err)

			response, err := GetDatabase(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Database.ID, response.Database.ID)
				assert.Equal(t, tc.expected.response.Database.Name, response.Database.Name)
			}
		})
	}
}

func TestCreateDatabase(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *CreateDatabaseResponse
		err      error
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDatabases []*model.Databases
		params            *CreateDatabaseParams
		expected          expected
	}{
		{
			scenario: "Will create and return a database",
			userData: &identity.UserData{ID: 1},
			params: &CreateDatabaseParams{
				Name: "test",
			},
			expected: expected{
				response: &CreateDatabaseResponse{
					Database: convert.DatabasePayload{
						Name: "test",
					},
				},
			},
		},
		{
			scenario: "Will throw an error when a database already exists",
			userData: &identity.UserData{ID: 1},
			params: &CreateDatabaseParams{
				Name: "test",
			},
			existingDatabases: []*model.Databases{
				{
					ID:        2,
					UserID:    1,
					Name:      "test",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "A database with name `test` already exists",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)

			err := insertDatabases(ctx, tc.existingDatabases)
			require.NoError(t, err)

			response, err := CreateDatabase(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Database.Name, response.Database.Name)
			}
		})
	}
}

func TestUpdateDatabase(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *UpdateDatabaseResponse
		err      error
	}

	validDatabase := &model.Databases{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDatabases []*model.Databases
		params            *UpdateDatabaseParams
		expected          expected
	}{
		{
			scenario: "Will update an existing database and return its data",
			userData: &identity.UserData{ID: 1},
			params: &UpdateDatabaseParams{
				ID:   validDatabase.ID,
				Name: "updated",
			},
			existingDatabases: []*model.Databases{validDatabase},
			expected: expected{
				response: &UpdateDatabaseResponse{
					Database: convert.DatabasePayload{
						ID:   validDatabase.ID,
						Name: "updated",
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the database does not exists",
			userData: &identity.UserData{ID: 1},
			params: &UpdateDatabaseParams{
				ID:   -1,
				Name: "updated",
			},
			existingDatabases: []*model.Databases{validDatabase},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
		{
			scenario: "Will throw an error when a database with this name already exists",
			userData: &identity.UserData{ID: 1},
			params: &UpdateDatabaseParams{
				ID:   validDatabase.ID,
				Name: validDatabase.Name,
			},
			existingDatabases: []*model.Databases{validDatabase},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "A database with name `test` already exists",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)

			err := insertDatabases(ctx, tc.existingDatabases)
			require.NoError(t, err)

			response, err := UpdateDatabase(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Database.Name, response.Database.Name)
			}
		})
	}
}

func TestDeleteDatabase(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *DeleteDatabaseResponse
		err      error
	}

	validDatabase := &model.Databases{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDatabases []*model.Databases
		params            *DeleteDatabaseParams
		expected          expected
	}{
		{
			scenario: "Will delete an existing database and return its data",
			userData: &identity.UserData{ID: 1},
			params: &DeleteDatabaseParams{
				ID: validDatabase.ID,
			},
			existingDatabases: []*model.Databases{validDatabase},
			expected: expected{
				response: &DeleteDatabaseResponse{
					Database: convert.DatabasePayload{
						ID:   validDatabase.ID,
						Name: validDatabase.Name,
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the database does not exists",
			userData: &identity.UserData{ID: 1},
			params: &DeleteDatabaseParams{
				ID: -1,
			},
			existingDatabases: []*model.Databases{validDatabase},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)

			err := insertDatabases(ctx, tc.existingDatabases)
			require.NoError(t, err)

			response, err := DeleteDatabase(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Database.Name, response.Database.Name)
			}
		})
	}
}
