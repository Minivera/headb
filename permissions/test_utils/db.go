package test_utils

import (
	"context"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

func Cleanup(ctx context.Context) error {
	query := `
		TRUNCATE permissions;
	`

	_, err := sqldb.Exec(ctx, query)
	if err != nil {
		log.WithError(err).Error("Could not clean db")
	}
	return err
}
