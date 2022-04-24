package helpers

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/qrm"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models"
	"encore.app/content/models/generated/content/public/model"
)

// GetDocument gets a document from a document ID and a user ID and returns a valid encore error
// if the document could not be fetched.
func GetDocument(ctx context.Context, documentID uuid.UUID, userID uuid.UUID) (*model.Documents, error) {
	document, err := models.GetDocumentByUser(ctx, documentID, userID)
	if errors.Is(err, qrm.ErrNoRows) {
		log.WithError(err).Warning("Could not find document by ID")
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Could not find document",
		}
	} else if err != nil {
		log.WithError(err).Error("Could not find document")
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Could not find document, unknown error",
		}
	}

	return document, nil
}
