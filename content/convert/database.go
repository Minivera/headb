package convert

import (
	"time"

	"encore.app/content/models/generated/content/public/model"
)

// DatabasePayload is an API safe version of a database.
type DatabasePayload struct {
	// The database unique identifier
	ID int64

	// The database unique name
	Name      string
	UpdatedAt time.Time
	CreatedAt time.Time
}

// DatabaseModelToPayload converts a database representation of a Database
// to an API safe version.
func DatabaseModelToPayload(database *model.Databases) DatabasePayload {
	return DatabasePayload{
		ID:        database.ID,
		Name:      database.Name,
		UpdatedAt: database.UpdatedAt,
		CreatedAt: database.CreatedAt,
	}
}

// DatabaseModelsToPayloads converts multiple database models to their API save versions
// using DatabaseModelToPayload.
func DatabaseModelsToPayloads(databases []*model.Databases) []DatabasePayload {
	converted := make([]DatabasePayload, len(databases))
	for i, database := range databases {
		converted[i] = DatabaseModelToPayload(database)
	}

	return converted
}
