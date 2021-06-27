package convert

import (
	"time"

	"encore.app/content/models"
)

type CollectionPayload struct {
	// The collection unique identifier
	ID uint64

	// The collection unique name
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

func CollectionModelToPayload(collection *models.Collection) CollectionPayload {
	return CollectionPayload{
		ID:        collection.ID,
		Name:      collection.Name,
		UpdatedAt: collection.UpdatedAt,
		CreatedAt: collection.CreatedAt,
	}
}

func CollectionModelsToPayloads(collections []*models.Collection) []CollectionPayload {
	converted := make([]CollectionPayload, len(collections))
	for i, collection := range collections {
		converted[i] = CollectionModelToPayload(collection)
	}

	return converted
}
