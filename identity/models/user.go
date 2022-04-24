package models

import (
	"context"

	"encore.dev/types/uuid"
	"github.com/go-jet/jet/v2/postgres"
	log "github.com/sirupsen/logrus"

	"encore.app/identity/models/generated/identity/public/model"
	"encore.app/identity/models/generated/identity/public/table"
)

// NewPendingUser creates a new User struct with all fields at their default values,
// setting the status to "pending".
func NewPendingUser() *model.Users {
	return &model.Users{
		Status: model.UserStatus_Pending,
	}
}

// GetUserByID fetches a user with an integer ID
func GetUserByID(ctx context.Context, id uuid.UUID) (*model.Users, error) {
	return GetUserBy(ctx, table.Users.ID.EQ(postgres.UUID(id)))
}

// GetUserByUniqueID fetches a user with the string unique ID given by the
// OAuth provider.
func GetUserByUniqueID(ctx context.Context, uniqueID string) (*model.Users, error) {
	return GetUserBy(ctx, table.Users.UniqueID.EQ(postgres.String(uniqueID)))
}

// GetUserBy is a utility function that fetches a user using a specific boolean expression.
// It returns a fully loaded user if it could be found or nil on an error.
func GetUserBy(ctx context.Context, expression postgres.BoolExpression) (*model.Users, error) {
	statement := postgres.SELECT(
		table.Users.ID,
		table.Users.Username,
		table.Users.Token,
		table.Users.UniqueID,
		table.Users.Status,
		table.Users.UpdatedAt,
		table.Users.CreatedAt,
	).FROM(table.Users).WHERE(expression).LIMIT(1)

	user := model.Users{}
	err := statement.QueryContext(ctx, db, &user)
	if err != nil {
		log.WithError(err).Error("Could not query user")
		return nil, err
	}

	return &user, nil
}

// SaveUser saves the data of the user it used on.
func SaveUser(ctx context.Context, user *model.Users) error {
	if user.ID == uuid.Nil {
		query, args := table.Users.INSERT(
			table.Users.Username,
			table.Users.Token,
			table.Users.UniqueID,
			table.Users.Status,
		).VALUES(
			user.Username,
			user.Token,
			user.UniqueID,
			user.Status,
		).RETURNING(
			table.Users.ID,
			table.Users.Username,
			table.Users.Token,
			table.Users.UniqueID,
			table.Users.Status,
			table.Users.UpdatedAt,
			table.Users.CreatedAt,
		).Sql()

		err := db.
			QueryRowContext(ctx, query, args...).
			Scan(
				&user.ID,
				&user.Username,
				&user.Token,
				&user.UniqueID,
				&user.Status,
				&user.UpdatedAt,
				&user.CreatedAt,
			)

		if err != nil {
			log.WithError(err).Error("Could not insert or update user")
			return err
		}

		return nil
	}

	// Use the non type safe version to allow for nulls
	query, args := table.Users.UPDATE(
		table.Users.Username,
		table.Users.Token,
		table.Users.UniqueID,
		table.Users.Status,
	).SET(
		user.Username,
		user.Token,
		user.UniqueID,
		user.Status,
	).WHERE(
		table.Users.ID.EQ(postgres.UUID(user.ID)),
	).RETURNING(
		table.Users.ID,
		table.Users.Username,
		table.Users.Token,
		table.Users.UniqueID,
		table.Users.Status,
		table.Users.UpdatedAt,
		table.Users.CreatedAt,
	).Sql()

	err := db.
		QueryRowContext(ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.Token, &user.UniqueID, &user.Status, &user.UpdatedAt, &user.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update user")
		return err
	}

	return nil
}

// DeleteUser deletes the user is it called on.
func DeleteUser(ctx context.Context, user *model.Users) error {
	query, args := table.Users.
		DELETE().
		WHERE(table.Users.ID.EQ(postgres.UUID(user.ID))).
		RETURNING(table.Users.ID).
		Sql()

	deletedID := uuid.Nil
	err := db.QueryRowContext(ctx, query, args...).Scan(&deletedID)

	if err != nil || deletedID == uuid.Nil {
		log.WithError(err).Error("Could not delete user")
		return err
	}

	return nil
}
