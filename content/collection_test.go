package content

import (
	"context"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	secondUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingDatabase := &model.Databases{
		ID:        newUUID,
		Name:      "test",
		UserID:    newUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollections := []*model.Collections{
		{
			ID:         newUUID,
			DatabaseID: existingDatabase.ID,
			Name:       "test",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         secondUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				DatabaseID: badUUID,
			},
			userCan: test_utils.StringPointer("read"),
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      newUUID,
					DatabaseID: &existingDatabase.ID,
					UserID:     newUUID,
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	secondUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	type expected struct {
		response *GetCollectionResponse
		err      error
	}

	existingDatabase := &model.Databases{
		ID:        newUUID,
		Name:      "test",
		UserID:    newUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollections := []*model.Collections{
		{
			ID:         newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan:             test_utils.StringPointer("read"),
			params:              &GetCollectionParams{ID: secondUUID},
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
				ID:    badUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      newUUID,
					DatabaseID: &existingDatabase.ID,
					UserID:     newUUID,
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	secondUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingDatabase := &model.Databases{
		ID:        newUUID,
		Name:      "test",
		UserID:    newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateCollectionParams{
				DatabaseID: badUUID,
				Name:       "test",
			},
			existingCollections: []*model.Collections{
				{
					ID:         secondUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateCollectionParams{
				DatabaseID: existingDatabase.ID,
				Name:       "test",
			},
			existingCollections: []*model.Collections{
				{
					ID:         secondUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      newUUID,
					DatabaseID: &existingDatabase.ID,
					UserID:     newUUID,
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	secondUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingDatabase := &model.Databases{
		ID:        newUUID,
		Name:      "test",
		UserID:    newUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollection := &model.Collections{
		ID:         secondUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateCollectionParams{
				ID:   badUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      newUUID,
					DatabaseID: &existingDatabase.ID,
					UserID:     newUUID,
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

	newUUID, err := uuid.NewV4()
	require.NoError(t, err)

	secondUUID, err := uuid.NewV4()
	require.NoError(t, err)

	badUUID, err := uuid.NewV4()
	require.NoError(t, err)

	existingDatabase := &model.Databases{
		ID:        newUUID,
		Name:      "test",
		UserID:    newUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validCollection := &model.Collections{
		ID:         secondUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &DeleteCollectionParams{
				ID: badUUID,
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
				ID:    newUUID,
				KeyID: newUUID,
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
			ctx := auth.WithContext(context.Background(), auth.UID(tc.userData.ID.String()), tc.userData)
			defer test_utils.Cleanup(ctx)
			defer test_utils_permissions.Cleanup(ctx)

			err := insertDatabases(ctx, []*model.Databases{existingDatabase})
			require.NoError(t, err)

			err = insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			if tc.userCan != nil {
				_, err := permissions.AddPermissionSet(ctx, &permissions.AddPermissionSetParams{
					KeyID:      newUUID,
					DatabaseID: &existingDatabase.ID,
					UserID:     newUUID,
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
