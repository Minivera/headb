package content

import (
	"context"

	"encore.app/content/convert"
	"encore.app/content/internal"
	"encore.dev/types/uuid"
)

// StartTransactionParams is the info needed to start a transaction on a user
type StartTransactionParams struct {
	// The unique identifier for the database to open a transaction for
	DatabaseID uuid.UUID
}

// StartTransactionResponse is the transaction to be used when creating content in this database
type StartTransactionResponse struct {
	// The unique identifier of the transaction, all requests to the database identified with this
	// key will run on the transaction.
	TransactionKey uuid.UUID
}

// StartTransaction starts a transaction on a database. A transaction is an isolated version of the database
// where all actions run on the data from that database and the transaction. All actions taken outside that
// transaction do not see the changes from that transaction until it is committed. A transaction can also be rolled back,
// which deletes it without applying it. Transactions are automatically rolled back after 10 minutes.
//encore:api auth
func StartTransaction(ctx context.Context, params *StartTransactionParams) (*StartTransactionResponse, error) {
	transaction, err := internal.StartTransaction(ctx, params.DatabaseID)
	if err != nil {
		return nil, err
	}

	return &StartTransactionResponse{
		TransactionKey: transaction.ID,
	}, nil
}

// CommitTransactionParams is the parameters for committing a transaction to a database.
type CommitTransactionParams struct {
	// The unique identifier of the database
	TransactionKey uuid.UUID
}

// CommitTransactionResponse is the result of a transaction commit
type CommitTransactionResponse struct {
	// The committed database
	Database convert.DatabasePayload
}

// CommitTransaction Finds a transaction by key and commits all its changes, deleting it in the process.
// A commit may fail, which will still trigger the transaction delete, but will not apply the changes.
//encore:api auth
func CommitTransaction(ctx context.Context, params *CommitTransactionParams) (*CommitTransactionResponse, error) {
	database, err := internal.CommitTransaction(ctx, params.TransactionKey)
	if err != nil {
		return nil, err
	}

	return &CommitTransactionResponse{
		Database: database,
	}, nil
}

// RollbackTransactionParams is the parameters for rollback-ing a transaction from a database.
type RollbackTransactionParams struct {
	// The unique identifier of the database
	TransactionKey uuid.UUID
}

// RollbackTransactionResponse is the result of a transaction rollback
type RollbackTransactionResponse struct {
	// The unique identifier of the transaction, all requests to the database identified with this
	// key will run on the transaction.
	TransactionKey uuid.UUID
}

// RollbackTransaction Finds a transaction by key and rollbacks all its changes, keeping the transaction alive,
// but deleting all its changes. The transaction can be reused
//encore:api auth
func RollbackTransaction(ctx context.Context, params *RollbackTransactionParams) (*RollbackTransactionResponse, error) {
	transaction, err := internal.RollbackTransaction(ctx, params.TransactionKey)
	if err != nil {
		return nil, err
	}

	return &RollbackTransactionResponse{
		TransactionKey: transaction.ID,
	}, nil
}
