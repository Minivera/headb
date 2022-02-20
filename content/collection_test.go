package content

import (
	"context"
	"strconv"
	"testing"
	"time"

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
	"encore.app/permissions"
	test_utils_permissions "encore.app/permissions/test_utils"
	test_utils2 "encore.app/test_utils"
)

func compareCollections(t *testing.T, expected, actual []convert.CollectionPayload) {
	assert.Len(t, actual, len(expected))
	for _, collection := range expected {
		var found *convert.CollectionPayload
		for _, lookup := range actual {
			if lookup.ID == collection.ID {
				found = &lookup
				break
			}
		}

		require.NotNilf(t, found, "Could not find collection with ID %d in actual", collection.ID)

		assert.Equal(t, collection.Name, found.Name)
	}
}

func insertCollections(ctx context.Context, collections []*model.Collections) error {
	for _, collection := range collections {
		query, args := table.Collections.INSERT(
			table.Collections.ID,
			table.Collections.Name,
			table.Collections.DatabaseID,
			table.Collections.UpdatedAt,
			table.Collections.CreatedAt,
		).VALUES(
			collection.ID,
			collection.Name,
			collection.DatabaseID,
			collection.UpdatedAt,
			collection.CreatedAt,
		).Sql()

		_, err := sqldb.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestListCollections(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *ListCollectionsResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        1,
		Name:      "test",
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollections := []*model.Collections{
		{
			ID:         2,
			DatabaseID: existingDatabase.ID,
			Name:       "test",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         3,
			DatabaseID: existingDatabase.ID,
			Name:       "test2",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		userCan             *string
		params              *ListCollectionsParams
		existingCollections []*model.Collections
		expected            expected
	}{
		{
			scenario: "Returns a list of collections owned by a database",
			params: &ListCollectionsParams{
				DatabaseID: existingDatabase.ID,
			},
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan:             test_utils.StringPointer("read"),
			existingCollections: validCollections,
			expected: expected{
				response: &ListCollectionsResponse{
					Collections: convert.CollectionModelsToPayloads(validCollections),
				},
			},
		},
		{
			scenario: "Returns empty when the database has no collections",
			params: &ListCollectionsParams{
				DatabaseID: existingDatabase.ID,
			},
			userCan: test_utils.StringPointer("read"),
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			existingCollections: []*model.Collections{},
			expected: expected{
				response: &ListCollectionsResponse{
					Collections: []convert.CollectionPayload{},
				},
			},
		},
		{
			scenario: "Fails when the database to fetch collections from doesn't exist",
			params: &ListCollectionsParams{
				DatabaseID: -1,
			},
			userCan: test_utils.StringPointer("read"),
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
		{
			scenario: "Fails when the key cannot read the database",
			params: &ListCollectionsParams{
				DatabaseID: existingDatabase.ID,
			},
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key doesn't have the ability to read the database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      1,
					DatabaseID: &existingDatabase.ID,
					UserID:     1,
					Role:       *tc.userCan,
				})
				require.NoError(t, err)
			}

			response, err := ListCollections(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				compareCollections(t, tc.expected.response.Collections, response.Collections)
			}
		})
	}
}

func TestGetCollection(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *GetCollectionResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        1,
		Name:      "test",
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollections := []*model.Collections{
		{
			ID:         2,
			DatabaseID: existingDatabase.ID,
			Name:       "test",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		userCan             *string
		params              *GetCollectionParams
		existingCollections []*model.Collections
		expected            expected
	}{
		{
			scenario: "Returns a collection by ID, owned by a user",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan:             test_utils.StringPointer("read"),
			params:              &GetCollectionParams{ID: validCollections[0].ID},
			existingCollections: validCollections,
			expected: expected{
				response: &GetCollectionResponse{
					Collection: convert.CollectionModelToPayload(validCollections[0]),
				},
			},
		},
		{
			scenario: "Returns an error when the collection is not found",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan:             test_utils.StringPointer("read"),
			params:              &GetCollectionParams{ID: 3},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Returns an error when the user does not own the collection",
			userData: &identity.UserData{
				ID:    2,
				KeyID: 1,
			},
			userCan:             test_utils.StringPointer("read"),
			params:              &GetCollectionParams{ID: validCollections[0].ID},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Fails when the key cannot read the database",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			params:              &GetCollectionParams{ID: validCollections[0].ID},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key doesn't have the ability to read the database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      1,
					DatabaseID: &existingDatabase.ID,
					UserID:     1,
					Role:       *tc.userCan,
				})
				require.NoError(t, err)
			}

			response, err := GetCollection(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Collection.ID, response.Collection.ID)
				assert.Equal(t, tc.expected.response.Collection.Name, response.Collection.Name)
			}
		})
	}
}

func TestCreateCollection(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *CreateCollectionResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        1,
		Name:      "test",
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		userCan             *string
		existingCollections []*model.Collections
		params              *CreateCollectionParams
		expected            expected
	}{
		{
			scenario: "Will create and return a collection",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateCollectionParams{
				DatabaseID: existingDatabase.ID,
				Name:       "test",
			},
			expected: expected{
				response: &CreateCollectionResponse{
					Collection: convert.CollectionPayload{
						Name: "test",
					},
				},
			},
		},
		{
			scenario: "Will throw an error when a the database cannot be found",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateCollectionParams{
				DatabaseID: -1,
				Name:       "test",
			},
			existingCollections: []*model.Collections{
				{
					ID:         2,
					DatabaseID: existingDatabase.ID,
					Name:       "test",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find database",
				},
			},
		},
		{
			scenario: "Will throw an error when a collection already exists",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateCollectionParams{
				DatabaseID: existingDatabase.ID,
				Name:       "test",
			},
			existingCollections: []*model.Collections{
				{
					ID:         2,
					DatabaseID: existingDatabase.ID,
					Name:       "test",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "A collection with name `test` already exists in this database",
				},
			},
		},
		{
			scenario: "Will throw an error fi the user cannot write to the database",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			params: &CreateCollectionParams{
				DatabaseID: existingDatabase.ID,
				Name:       "test",
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key doesn't have the ability to write to the database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      1,
					DatabaseID: &existingDatabase.ID,
					UserID:     1,
					Role:       *tc.userCan,
				})
				require.NoError(t, err)
			}

			response, err := CreateCollection(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Collection.Name, response.Collection.Name)
			}
		})
	}
}

func TestUpdateCollection(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *UpdateCollectionResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        1,
		Name:      "test",
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollection := &model.Collections{
		ID:         2,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		userCan             *string
		existingCollections []*model.Collections
		params              *UpdateCollectionParams
		expected            expected
	}{
		{
			scenario: "Will update an existing collection and return its data",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateCollectionParams{
				ID:   validCollection.ID,
				Name: "updated",
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				response: &UpdateCollectionResponse{
					Collection: convert.CollectionPayload{
						ID:   validCollection.ID,
						Name: "updated",
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the collection does not exists",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateCollectionParams{
				ID:   -1,
				Name: "updated",
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Will throw an error when a collection with this name already exists",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateCollectionParams{
				ID:   validCollection.ID,
				Name: validCollection.Name,
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "A collection with name `test` already exists in this database",
				},
			},
		},
		{
			scenario: "Will fail if the key cannot write to the database",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			params: &UpdateCollectionParams{
				ID:   validCollection.ID,
				Name: "updated",
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key doesn't have the ability to write to the database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      1,
					DatabaseID: &existingDatabase.ID,
					UserID:     1,
					Role:       *tc.userCan,
				})
				require.NoError(t, err)
			}

			response, err := UpdateCollection(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Collection.Name, response.Collection.Name)
			}
		})
	}
}

func TestDeleteCollection(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *DeleteCollectionResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        1,
		Name:      "test",
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollection := &model.Collections{
		ID:         2,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		userCan             *string
		existingCollections []*model.Collections
		params              *DeleteCollectionParams
		expected            expected
	}{
		{
			scenario: "Will delete an existing collection and return its data",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &DeleteCollectionParams{
				ID: validCollection.ID,
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				response: &DeleteCollectionResponse{
					Collection: convert.CollectionPayload{
						ID:   validCollection.ID,
						Name: validCollection.Name,
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the collection does not exists",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			userCan: test_utils.StringPointer("write"),
			params: &DeleteCollectionParams{
				ID: -1,
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Will fail if the key cannot write to the database",
			userData: &identity.UserData{
				ID:    1,
				KeyID: 1,
			},
			params: &DeleteCollectionParams{
				ID: validCollection.ID,
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.PermissionDenied,
					Message: "API key doesn't have the ability to write to the database",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := auth.WithContext(context.Background(), auth.UID(strconv.FormatInt(tc.userData.ID, 10)), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      1,
					DatabaseID: &existingDatabase.ID,
					UserID:     1,
					Role:       *tc.userCan,
				})
				require.NoError(t, err)
			}

			response, err := DeleteCollection(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Collection.Name, response.Collection.Name)
			}
		})
	}
}
