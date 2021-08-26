package content

import (
	"context"
	"testing"
	"time"

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
			table.Collections.UserID,
			table.Collections.UpdatedAt,
			table.Collections.CreatedAt,
		).VALUES(
			collection.ID,
			collection.Name,
			collection.UserID,
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

	validCollections := []*model.Collections{
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
		scenario            string
		userData            *identity.UserData
		existingCollections []*model.Collections
		expected            expected
	}{
		{
			scenario:            "Returns a list of collections owned by a user",
			userData:            &identity.UserData{ID: 1},
			existingCollections: validCollections,
			expected: expected{
				response: &ListCollectionsResponse{
					Collections: convert.CollectionModelsToPayloads(validCollections),
				},
			},
		},
		{
			scenario:            "Returns empty when the user owns no collections",
			userData:            &identity.UserData{ID: 2},
			existingCollections: validCollections,
			expected: expected{
				response: &ListCollectionsResponse{
					Collections: []convert.CollectionPayload{},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			response, err := listCollections(ctx, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollections := []*model.Collections{
		{
			ID:        2,
			UserID:    1,
			Name:      "test",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		params              *GetCollectionParams
		existingCollections []*model.Collections
		expected            expected
	}{
		{
			scenario:            "Returns a collection by ID, owned by a user",
			userData:            &identity.UserData{ID: 1},
			params:              &GetCollectionParams{ID: validCollections[0].ID},
			existingCollections: validCollections,
			expected: expected{
				response: &GetCollectionResponse{
					Collection: convert.CollectionModelToPayload(validCollections[0]),
				},
			},
		},
		{
			scenario:            "Returns an error when the collection is not found",
			userData:            &identity.UserData{ID: 1},
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
			scenario:            "Returns an error when the user does not own the collection",
			userData:            &identity.UserData{ID: 2},
			params:              &GetCollectionParams{ID: validCollections[0].ID},
			existingCollections: validCollections,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			response, err := getCollection(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		existingCollections []*model.Collections
		params              *CreateCollectionParams
		expected            expected
	}{
		{
			scenario: "Will create and return a collection",
			userData: &identity.UserData{ID: 1},
			params: &CreateCollectionParams{
				Name: "test",
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
			scenario: "Will throw an error when a collection already exists",
			userData: &identity.UserData{ID: 1},
			params: &CreateCollectionParams{
				Name: "test",
			},
			existingCollections: []*model.Collections{
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
					Message: "A collection with name `test` already exists for this user",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			response, err := createCollection(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		existingCollections []*model.Collections
		params              *UpdateCollectionParams
		expected            expected
	}{
		{
			scenario: "Will update an existing collection and return its data",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
			params: &UpdateCollectionParams{
				ID:   validCollection.ID,
				Name: validCollection.Name,
			},
			existingCollections: []*model.Collections{validCollection},
			expected: expected{
				err: &errs.Error{
					Code:    errs.AlreadyExists,
					Message: "A collection with name `test` already exists for this user",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			response, err := updateCollection(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario            string
		userData            *identity.UserData
		existingCollections []*model.Collections
		params              *DeleteCollectionParams
		expected            expected
	}{
		{
			scenario: "Will delete an existing collection and return its data",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
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
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, tc.existingCollections)
			require.NoError(t, err)

			response, err := deleteCollection(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.response.Collection.Name, response.Collection.Name)
			}
		})
	}
}
