package content

import (
	"context"
	"encoding/json"
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
	validDocuments := []*model.Documents{
		{
			ID:           3,
			CollectionID: validCollections[0].ID,
			Content:      `{"foo": "bar"}`,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           4,
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
		params            *ListDocumentsParams
		existingDocuments []*model.Documents
		expected          expected
	}{
		{
			scenario: "Returns a list of documents owned by a user",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
			params: &ListDocumentsParams{
				CollectionID: validCollections[0].ID,
			},
			existingDocuments: []*model.Documents{
				{
					ID:           3,
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
			userData: &identity.UserData{ID: 1},
			params: &ListDocumentsParams{
				CollectionID: -1,
			},
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

			err = insertCollections(ctx, validCollections)
			require.NoError(t, err)

			err := insertDocuments(ctx, tc.existingDocuments)
			require.NoError(t, err)

			response, err := listDocuments(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	validDocuments := []*model.Documents{
		{
			ID:           3,
			CollectionID: 2,
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
		params            *GetDocumentParams
		existingDocuments []*model.Documents
		expected          expected
	}{
		{
			scenario:          "Returns a document by ID, owned by a user",
			userData:          &identity.UserData{ID: 1},
			params:            &GetDocumentParams{ID: validDocuments[0].ID},
			existingDocuments: validDocuments,
			expected: expected{
				response: &GetDocumentResponse{
					Document: documentPayloads[0],
				},
			},
		},
		{
			scenario:          "Returns an error when the document is not found",
			userData:          &identity.UserData{ID: 1},
			params:            &GetDocumentParams{ID: -1},
			existingDocuments: validDocuments,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
		{
			scenario:          "Returns an error when the user does not own the document",
			userData:          &identity.UserData{ID: -1},
			params:            &GetDocumentParams{ID: validDocuments[0].ID},
			existingDocuments: validDocuments,
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err = insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err := insertDocuments(ctx, tc.existingDocuments)
			require.NoError(t, err)

			response, err := getDocument(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tcs := []struct {
		scenario string
		userData *identity.UserData
		params   *CreateDocumentParams
		expected expected
	}{
		{
			scenario: "Will create and return a document",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
			params: &CreateDocumentParams{
				CollectionID: -1,
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
			userData: &identity.UserData{ID: 1},
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
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			response, err := createDocument(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	validDocument := &model.Documents{
		ID:           2,
		CollectionID: validCollection.ID,
		Content:      `{"foo": "bar"}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDocuments []*model.Documents
		params            *UpdateDocumentParams
		expected          expected
	}{
		{
			scenario: "Will update an existing document and return its data",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
			params: &UpdateDocumentParams{
				ID:      -1,
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
			userData: &identity.UserData{ID: 1},
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
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
			require.NoError(t, err)

			response, err := updateDocument(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
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

	validCollection := &model.Collections{
		ID:        2,
		UserID:    1,
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	validDocument := &model.Documents{
		ID:           2,
		CollectionID: validCollection.ID,
		Content:      `{"foo": "bar"}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tcs := []struct {
		scenario          string
		userData          *identity.UserData
		existingDocuments []*model.Documents
		params            *DeleteDocumentParams
		expected          expected
	}{
		{
			scenario: "Will delete an existing document and return its data",
			userData: &identity.UserData{ID: 1},
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
			userData: &identity.UserData{ID: 1},
			params: &DeleteDocumentParams{
				ID: -1,
			},
			existingDocuments: []*model.Documents{validDocument},
			expected: expected{
				err: &errs.Error{
					Code:    errs.NotFound,
					Message: "Could not find document",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.scenario, func(t *testing.T) {
			ctx := context.Background()
			defer test_utils.Cleanup(ctx)

			err := insertCollections(ctx, []*model.Collections{validCollection})
			require.NoError(t, err)

			err = insertDocuments(ctx, tc.existingDocuments)
			require.NoError(t, err)

			response, err := deleteDocument(ctx, tc.params, tc.userData)
			if err != nil {
				assert.Equal(t, tc.expected.err, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, string(tc.expected.response.Document.Content), string(response.Document.Content))
			}
		})
	}
}
