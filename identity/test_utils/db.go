package test_utils

import (
	"context"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

var db = sqldb.Named("identity").Stdlib()

func Cleanup(ctx context.Context) error {
	query := `
		TRUNCATE api_keys, users;
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		log.WithError(err).Error("Could not clean db")
	}
	return err
}
