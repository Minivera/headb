package content

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	log "github.com/sirupsen/logrus"
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

func compareDocuments(t *testing.T, expected, actual []convert.DocumentPayload) {
	assert.Len(t, actual, len(expected))
	for _, document := range expected {
		var found *convert.DocumentPayload
		for _, lookup := range actual {
			if lookup.ID == document.ID {
				found = &lookup
				break
			}
		}

		require.NotNilf(t, found, "Could not find document with ID %d in actual", document.ID)

		assert.Equal(t, string(document.Content), string(found.Content))
	}
}

func insertDocuments(ctx context.Context, documents []*model.Documents) error {
	for _, document := range documents {
		query, args := table.Documents.INSERT(
			table.Documents.ID,
			table.Documents.Content,
			table.Documents.CollectionID,
			table.Documents.UpdatedAt,
			table.Documents.CreatedAt,
		).VALUES(
			document.ID,
			document.Content,
			document.CollectionID,
			document.UpdatedAt,
			document.CreatedAt,
		).Sql()

		_, err := sqldb.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestListDocuments(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *ListDocumentsResponse
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
	validDocuments := []*model.Documents{
		{
			ID:           newUUID,
			CollectionID: validCollections[0].ID,
			Content:      `{"foo": "bar"}`,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           secondUUID,
			CollectionID: validCollections[0].ID,
			Content:      `{"foo": "bar2"}`,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	documentPayloads, err := convert.DocumentModelsToPayloads(validDocuments)
	require.NoError(t, err)

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		userCan           *string
		params            *ListDocumentsParams
		existingDocuments []*model.Documents
		expected          expected
	}{
		{
			scenario: "Returns a list of documents owned by a user",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("read"),
			params: &ListDocumentsParams{
				CollectionID: validCollections[0].ID,
			},
			existingDocuments: validDocuments,
			expected: expected{
				response: &ListDocumentsResponse{
					Documents: documentPayloads,
				},
			},
		},
		{
			scenario: "Returns empty when the collection owns no documents",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("read"),
			params: &ListDocumentsParams{
				CollectionID: validCollections[0].ID,
			},
			existingDocuments: []*model.Documents{
				{
					ID:           badUUID,
					CollectionID: validCollections[1].ID,
					Content:      `{"foo": "bar"}`,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
			expected: expected{
				response: &ListDocumentsResponse{
					Documents: []convert.DocumentPayload{},
				},
			},
		},
		{
			scenario: "Throws an error when the collection does not exists",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("read"),
			params: &ListDocumentsParams{
				CollectionID: badUUID,
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Fails when the user cannot read the database",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			params: &ListDocumentsParams{
				CollectionID: validCollections[0].ID,
			},
			existingDocuments: validDocuments,
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

			err = insertCollections(ctx, validCollections)
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
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

			response, err := ListDocuments(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				compareDocuments(t, tc.expected.response.Documents, response.Documents)
			}
		})
	}
}

func TestGetDocument(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *GetDocumentResponse
		err      error
	}

	newUUID, err := uuid.NewV4()
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
		ID:         newUUID,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	validDocuments := []*model.Documents{
		{
			ID:           newUUID,
			CollectionID: newUUID,
			Content:      `{"foo": "bar"}`,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	documentPayloads, err := convert.DocumentModelsToPayloads(validDocuments)
	require.NoError(t, err)

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		userCan           *string
		params            *GetDocumentParams
		existingDocuments []*model.Documents
		expected          expected
	}{
		{
			scenario: "Returns a document by ID, owned by a user",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan:           test_utils.StringPointer("read"),
			params:            &GetDocumentParams{ID: validDocuments[0].ID},
			existingDocuments: validDocuments,
			expected: expected{
				response: &GetDocumentResponse{
					Document: documentPayloads[0],
				},
			},
		},
		{
			scenario: "Returns an error when the document is not found",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan:           test_utils.StringPointer("read"),
			params:            &GetDocumentParams{ID: badUUID},
			existingDocuments: validDocuments,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
		{
			scenario: "Returns an error when the user does not own the document",
			userData: &identity.UserData{
				ID:    badUUID,
				KeyID: newUUID,
			},
			userCan:           test_utils.StringPointer("read"),
			params:            &GetDocumentParams{ID: validDocuments[0].ID},
			existingDocuments: validDocuments,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
		{
			scenario: "Returns an error when the key cannot access the database",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			params:            &GetDocumentParams{ID: validDocuments[0].ID},
			existingDocuments: validDocuments,
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

			err = insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
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

			response, err := GetDocument(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.response.Document.ID, response.Document.ID)
				assert.Equal(t, tc.expected.response.Document.Content, response.Document.Content)
			}
		})
	}
}

func TestCreateDocument(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *CreateDocumentResponse
		err      error
	}

	newUUID, err := uuid.NewV4()
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
		ID:         newUUID,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tcs := []struct {
		scenario string
		userData *identity.UserData
		userCan  *string
		params   *CreateDocumentParams
		expected expected
	}{
		{
			scenario: "Will create and return a document",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateDocumentParams{
				CollectionID: validCollection.ID,
				Content:      json.RawMessage(`{"foo": "bar"}`),
			},
			expected: expected{
				response: &CreateDocumentResponse{
					Document: convert.DocumentPayload{
						Content: json.RawMessage(`"{\"foo\": \"bar\"}"`),
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
			params: &CreateDocumentParams{
				CollectionID: badUUID,
				Content:      json.RawMessage(`{"foo": "bar"}`),
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find collection",
				},
			},
		},
		{
			scenario: "Will throw an error when the content is not JSON",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &CreateDocumentParams{
				CollectionID: validCollection.ID,
				Content:      json.RawMessage("test"),
			},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Received JSON string for content was not valid",
				},
			},
		},
		{
			scenario: "Will return an error when the key cannot write to the database",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			params: &CreateDocumentParams{
				CollectionID: validCollection.ID,
				Content:      json.RawMessage(`{"foo": "bar"}`),
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

			err = insertCollections(ctx, []*model.Collections{validCollection})
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

			response, err := CreateDocument(ctx, tc.params)
			fmt.Printf("%s: %v", tc.scenario, string(tc.params.Content))
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				log.Warningf("maybe nil %s: %v", tc.scenario, err)
				require.NoError(t, err)
				assert.Equal(t, string(tc.expected.response.Document.Content), string(response.Document.Content))
			}
		})
	}
}

func TestUpdateDocument(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *UpdateDocumentResponse
		err      error
	}

	newUUID, err := uuid.NewV4()
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
		ID:         newUUID,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	validDocument := &model.Documents{
		ID:           newUUID,
		CollectionID: validCollection.ID,
		Content:      `{"foo": "bar"}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		userCan           *string
		existingDocuments []*model.Documents
		params            *UpdateDocumentParams
		expected          expected
	}{
		{
			scenario: "Will update an existing document and return its data",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateDocumentParams{
				ID:      validDocument.ID,
				Content: json.RawMessage(`{"foo": "updated"}`),
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				response: &UpdateDocumentResponse{
					Document: convert.DocumentPayload{
						ID:      validDocument.ID,
						Content: json.RawMessage(`"{\"foo\": \"updated\"}"`),
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the document does not exists",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateDocumentParams{
				ID:      badUUID,
				Content: json.RawMessage(`{"foo": "updated"}`),
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
		{
			scenario: "Will throw an error when the JSON is not valid",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &UpdateDocumentParams{
				ID:      validDocument.ID,
				Content: json.RawMessage("test"),
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				err: &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "Received JSON string for content was not valid",
				},
			},
		},
		{
			scenario: "Will return an error when the key cannot write to the database",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			params: &UpdateDocumentParams{
				ID:      validDocument.ID,
				Content: json.RawMessage(`{"foo": "updated"}`),
			},
			existingDocuments: []*model.Documents{validDocument},
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

			err = insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
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

			response, err := UpdateDocument(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, string(tc.expected.response.Document.Content), string(response.Document.Content))
			}
		})
	}
}

func TestDeleteDocument(t *testing.T) {
	now := time.Now()

	type expected struct {
		response *DeleteDocumentResponse
		err      error
	}

	newUUID, err := uuid.NewV4()
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
		ID:         newUUID,
		DatabaseID: existingDatabase.ID,
		Name:       "test",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	validDocument := &model.Documents{
		ID:           newUUID,
		CollectionID: validCollection.ID,
		Content:      `{"foo": "bar"}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		userCan           *string
		existingDocuments []*model.Documents
		params            *DeleteDocumentParams
		expected          expected
	}{
		{
			scenario: "Will delete an existing document and return its data",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &DeleteDocumentParams{
				ID: validDocument.ID,
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				response: &DeleteDocumentResponse{
					Document: convert.DocumentPayload{
						ID:      validDocument.ID,
						Content: json.RawMessage(`"{\"foo\": \"bar\"}"`),
					},
				},
			},
		},
		{
			scenario: "Will throw an error when the document does not exists",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			userCan: test_utils.StringPointer("write"),
			params: &DeleteDocumentParams{
				ID: badUUID,
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
		{
			scenario: "Will fail when the key cannot write to the database",
			userData: &identity.UserData{
				ID:    newUUID,
				KeyID: newUUID,
			},
			params: &DeleteDocumentParams{
				ID: validDocument.ID,
			},
			existingDocuments: []*model.Documents{validDocument},
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

			err = insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
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

			response, err := DeleteDocument(ctx, tc.params)
			if tc.expected.err != nil {
				test_utils2.CompareErrors(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				assert.Equal(t, string(tc.expected.response.Document.Content), string(response.Document.Content))
			}
		})
	}
}
