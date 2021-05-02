package convert

import (
	"encoding/json"
	"time"

	"encore.app/headb/models"
)

type DocumentPayload struct {
	// The document unique identifier
	ID uint64

	// The document content (Stringified JSON for now)
	Content   string
	UpdatedAt time.Time
	CreatedAt time.Time
}

func DocumentModelToPayload(document *models.Document) (DocumentPayload, error) {
	contentString, err := json.Marshal(document.Content)
	if err != nil {
		return DocumentPayload{}, err
	}

	return DocumentPayload{
		ID:        document.ID,
		Content:   string(contentString),
		UpdatedAt: document.UpdatedAt,
		CreatedAt: document.CreatedAt,
	}, nil
}

func DocumentModelsToPayloads(documents []*models.Document) ([]DocumentPayload, error) {
	converted := make([]DocumentPayload, len(documents))
	for i, document := range documents {
		var err error
		converted[i], err = DocumentModelToPayload(document)
		if err != nil {
			return nil, err
		}
	}

	return converted, nil
}
