package test_utils

import (
	"context"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

var db = sqldb.Named("permissions").Stdlib()

func Cleanup(ctx context.Context) error {
	query := `
		TRUNCATE permissions;
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("Could not clean db")
	}
	return err
}
