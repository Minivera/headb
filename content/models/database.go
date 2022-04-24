package models

import (
	"context"

	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/content/models/generated/content/public/model"
	"encore.app/content/models/generated/content/public/table"
)

var db = sqldb.Named("content").Stdlib()

// NewDatabase generates a new database structure from a name and the
// associated user ID.
func NewDatabase(name string, userID uuid.UUID) *model.Databases {
	return &model.Databases{
		Name:   name,
		UserID: userID,
	}
}

// ListDatabase lists all databases for a given user, it returns
// a nil database on an error.
func ListDatabase(ctx context.Context, userID uuid.UUID) ([]*model.Databases, error) {
	statement := postgres.SELECT(
		table.Databases.ID,
		table.Databases.Name,
		table.Databases.UserID,
		table.Databases.UpdatedAt,
		table.Databases.CreatedAt,
	).FROM(table.Databases).WHERE(
		table.Databases.UserID.EQ(postgres.UUID(userID)),
	)

	var databases []*model.Databases
	err := statement.QueryContext(ctx, db, &databases)
	if err != nil {
		log.WithError(err).Error("Could not query databases")
		return nil, err
	}

	return databases, nil
}

// GetDatabaseByID fetches a single database record given an ID and the associated
// user ID. Returns nil on an error.
func GetDatabaseByID(ctx context.Context, id, userID uuid.UUID) (*model.Databases, error) {
	statement := postgres.SELECT(
		table.Databases.ID,
		table.Databases.Name,
		table.Databases.UserID,
		table.Databases.UpdatedAt,
		table.Databases.CreatedAt,
	).FROM(
		table.Databases,
	).WHERE(
		table.Databases.ID.EQ(postgres.UUID(id)).
			AND(table.Databases.UserID.EQ(postgres.UUID(userID))),
	).LIMIT(1)

	database := model.Databases{}
	err := statement.QueryContext(ctx, db, &database)
	if err != nil {
		log.WithError(err).Errorf("Could not query database for id %d", id)
		return nil, err
	}

	return &database, nil
}

// ValidateDatabaseConstraint validates that no database with the same name exists
// for a single user.
func ValidateDatabaseConstraint(ctx context.Context, database *model.Databases) bool {
	query, args := postgres.SELECT(
		table.Databases.ID,
	).FROM(
		table.Databases,
	).WHERE(
		table.Databases.Name.EQ(postgres.String(database.Name)).
			AND(table.Databases.UserID.EQ(postgres.UUID(database.UserID))),
	).LIMIT(1).Sql()

	id := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == nil && id != uuid.Nil {
		log.Warning("Tried to save database, a database already exists for this name and user_id")
		return false
	}

	return true
}

// SaveDatabase saves the data of the database it used on. This method only saves
// the name and user ID from the struct and updates the timestamps. SaveDatabase will
// trigger an error if the constraints are not respected.
func SaveDatabase(ctx context.Context, database *model.Databases) error {
	if database.ID == uuid.Nil {
		query, args := table.Databases.INSERT(
			table.Databases.Name,
			table.Databases.UserID,
		).VALUES(
			database.Name,
			database.UserID,
		).RETURNING(
			table.Databases.ID,
			table.Databases.UpdatedAt,
			table.Databases.CreatedAt,
		).Sql()

		err := db.
			QueryRowContext(ctx, query, args...).
			Scan(&database.ID, &database.UpdatedAt, &database.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert database")
			return err
		}

		return nil
	}

	query, args := table.Databases.UPDATE().SET(
		table.Databases.Name.SET(postgres.String(database.Name)),
		table.Databases.UserID.SET(postgres.UUID(database.UserID)),
	).WHERE(
		table.Databases.ID.EQ(postgres.UUID(database.ID)),
	).RETURNING(
		table.Databases.ID,
		table.Databases.UpdatedAt,
		table.Databases.CreatedAt,
	).Sql()

	err := db.
		QueryRowContext(ctx, query, args...).
		Scan(&database.ID, &database.UpdatedAt, &database.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update database")
		return err
	}

	return nil
}

// DeleteDatabase deletes the database is it called on.
func DeleteDatabase(ctx context.Context, database *model.Databases) error {
	query, args := table.Databases.
		DELETE().
		WHERE(table.Databases.ID.EQ(postgres.UUID(database.ID))).
		RETURNING(table.Databases.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)
	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete database")
		return err
	}

	return nil
}
