package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

const (
	// UserStatusPending represents the status of a user that is still waiting authentication
	UserStatusPending = "pending"

	// UserStatusAccepted represents the status of a user whose auth was accepted.
	UserStatusAccepted = "accepted"

	// UserStatusDenied represents the status of a user whose auth was denied, either because
	// they did not accept it or because the OAuth provider denied it.
	UserStatusDenied = "denied"
)

// User is the struct representing a user from the database table.
type User struct {
	ID        uint64
	Username  string
	Token     string
	UniqueID  string
	Status    string
	UpdatedAt time.Time
	CreatedAt time.Time
}

// NewPendingUser creates a new User struct with all fields at their default values,
// setting the status to "pending".
func NewPendingUser() *User {
	return &User{
		Status: UserStatusPending,
	}
}

// GetUserByID fetches a user with an integer ID
func GetUserByID(ctx context.Context, ID uint64) (*User, error) {
	return GetUserBy(ctx, "id", strconv.FormatUint(ID, 10))
}

// GetUserByUniqueID fetches a user with the string unique ID given by the
// OAuth provider.
func GetUserByUniqueID(ctx context.Context, uniqueID string) (*User, error) {
	return GetUserBy(ctx, "unique_id", fmt.Sprintf("'%s'", uniqueID))
}

// GetUserBy is a utility function that fetches a user using a specific field and
// value. It returns a fully loaded user if it could be found or nil on an error.
func GetUserBy(ctx context.Context, field, val string) (*User, error) {
	userQuery := fmt.Sprintf(`
		SELECT
			id,
			username,
			token,
			unique_id,
			status,
			updated_at,
			created_at
		FROM
			"users"
		WHERE
			users.%s = $1
		LIMIT 1;
	`, field)

	user := User{}
	err := sqldb.
		QueryRow(ctx, userQuery, val).
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
		log.WithError(err).Errorf("Could not query user for %s %s", field, val)
		return nil, err
	}

	return &user, nil
}

// Save saves the data of the key it used on.
func (user *User) Save(ctx context.Context) error {
	if user.ID == 0 {
		userQuery := `
			INSERT INTO "users" (username, token, unique_id, status)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ON CONSTRAINT user_username_unique DO UPDATE SET token = $2, status = $4
		RETURNING id, username, token, unique_id, status, updated_at, created_at;
		`

		err := sqldb.
			QueryRow(ctx, userQuery, user.Username, user.Token, user.UniqueID, user.Status).
			Scan(&user.ID, &user.Username, &user.Token, &user.UniqueID, &user.Status, &user.UpdatedAt, &user.CreatedAt)

		if err != nil {
			log.WithError(err).Error("Could not insert or update user")
			return err
		}

		return nil
	}

	collectionQuery := `
		UPDATE "users"
		SET username = $1, token = $2, unique_id = $3, status = $4
		WHERE id = $5
		RETURNING id, username, token, unique_id, status, updated_at, created_at;
	`

	err := sqldb.
		QueryRow(ctx, collectionQuery, user.Username, user.Token, user.UniqueID, user.Status, user.ID).
		Scan(&user.ID, &user.Username, &user.Token, &user.UniqueID, &user.Status, &user.UpdatedAt, &user.CreatedAt)

	if err != nil {
		log.WithError(err).Error("Could not update user")
		return err
	}

	return nil
}

// Delete deletes the API key is it called on.
func (user *User) Delete(ctx context.Context) error {
	userQuery := `
		DELETE FROM "users"
		WHERE id = $1
		RETURNING id;
	`

	deletedID := 0
	err := sqldb.QueryRow(ctx, userQuery, user.ID).Scan(&deletedID)

	if err != nil || deletedID == 0 {
		log.WithError(err).Error("Could not delete user")
		return err
	}

	return nil
}
