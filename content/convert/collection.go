package convert

import (
	"time"

	"encore.app/content/models"
)

// CollectionPayload is an API safe version of a collection.
type CollectionPayload struct {
	// The collection unique identifier
	ID uint64

	// The collection unique name
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

// CollectionModelToPayload converts a database representation of a Collection
// to an API safe version.
func CollectionModelToPayload(collection *models.Collection) CollectionPayload {
	return CollectionPayload{
		ID:        collection.ID,
		Name:      collection.Name,
		UpdatedAt: collection.UpdatedAt,
		CreatedAt: collection.CreatedAt,
	}
}

// CollectionModelsToPayloads converts multiple collection models to their API save versions
// using CollectionModelToPayload.
func CollectionModelsToPayloads(collections []*models.Collection) []CollectionPayload {
	converted := make([]CollectionPayload, len(collections))
	for i, collection := range collections {
		converted[i] = CollectionModelToPayload(collection)
	}

	return converted
}
