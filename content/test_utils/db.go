package test_utils

import (
	"context"

	log "github.com/sirupsen/logrus"

	"encore.dev/storage/sqldb"
)

func Cleanup(ctx context.Context) error {
	query := `
		TRUNCATE documents, collections;
	`

	_, err := sqldb.Exec(ctx, query)
	if err != nil {
		log.WithError(err).Error("Could not clean db")
	}
	return err
}
