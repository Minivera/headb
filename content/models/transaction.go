package models

import (
	"context"
	"time"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

const TransactionExpirationTime = 10 * time.Minute

// NewTransaction generates a new transaction structure using the provided database ID and sets its
// expiration date to 10 minutes from now.
func NewTransaction(databaseID uuid.UUID) *model.Transactions {
	return &model.Transactions{
		DatabaseID: databaseID,
		ExpiresAt:  time.Now().Add(TransactionExpirationTime),
		Script: `BEGIN;
"{{.Sql}}
COMMIT;`,
	}
}

// GetTransactionByID fetches a single transaction record given an ID and the associated
// database ID. Returns nil on an error.
func GetTransactionByID(ctx context.Context, id, databaseID uuid.UUID) (*model.Transactions, error) {
	statement := postgres.SELECT(
		table.Transactions.ID,
		table.Transactions.Script,
		table.Transactions.DatabaseID,
		table.Transactions.ExpiresAt,
		table.Transactions.UpdatedAt,
		table.Transactions.CreatedAt,
	).FROM(
		table.Transactions,
	).WHERE(
		table.Transactions.ID.EQ(postgres.UUID(id)).
			AND(table.Transactions.DatabaseID.EQ(postgres.UUID(databaseID))),
	).LIMIT(1)

	transaction := model.Transactions{}
	err := statement.QueryContext(ctx, db, &transaction)
	if err != nil {
		log.WithError(err).Errorf("Could not query transaction for id %d", id)
		return nil, err
	}

	return &transaction, nil
}

// SaveTransaction saves the data of the transaction it used on. SaveTransaction will
// trigger an error if the constraints are not respected.
func SaveTransaction(ctx context.Context, transaction *model.Transactions) error {
	if transaction.ID == uuid.Nil {
		query, args := table.Collections.INSERT(
			table.Transactions.Script,
			table.Transactions.DatabaseID,
			table.Transactions.ExpiresAt,
		).VALUES(
			transaction.Script,
			transaction.DatabaseID,
			transaction.ExpiresAt,
		).RETURNING(
			table.Collections.ID,
			table.Collections.UpdatedAt,
			table.Collections.CreatedAt,
		).Sql()

		err := db.
			QueryRowContext(ctx, query, args...).
			Scan(&transaction.ID, &transaction.UpdatedAt, &transaction.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert transaction")
			return err
		}

		return nil
	}

	query, args := table.Transactions.UPDATE().SET(
		table.Transactions.Script.SET(postgres.String(transaction.Script)),
		table.Transactions.DatabaseID.SET(postgres.UUID(transaction.DatabaseID)),
		table.Transactions.ExpiresAt.SET(postgres.TimestampzT(transaction.ExpiresAt)),
	).WHERE(
		table.Transactions.ID.EQ(postgres.UUID(transaction.ID)),
	).RETURNING(
		table.Transactions.ID,
		table.Transactions.UpdatedAt,
		table.Transactions.CreatedAt,
	).Sql()

	err := db.
		QueryRowContext(ctx, query, args...).
		Scan(&transaction.ID, &transaction.UpdatedAt, &transaction.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update transaction")
		return err
	}

	return nil
}

// DeleteTransaction deletes the Transaction is it called on.
func DeleteTransaction(ctx context.Context, transaction *model.Transactions) error {
	query, args := table.Transactions.
		DELETE().
		WHERE(table.Transactions.ID.EQ(postgres.UUID(transaction.ID))).
		RETURNING(table.Transactions.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete transaction")
		return err
	}

	return nil
}
