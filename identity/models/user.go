package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"encore.dev/storage/sqldb"
	log "github.com/sirupsen/logrus"
)

type User struct {
	ID        uint64
	Username  string
	Token     string
	UpdatedAt time.Time
	CreatedAt time.Time
}

func NewUser(username string) *User {
	return &User{
		Username: username,
	}
}

func GetUserByID(ctx context.Context, ID uint64) (*User, error) {
	return GetUserBy(ctx, "id", strconv.FormatUint(ID, 10))
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return GetUserBy(ctx, "username", username)
}

func GetUserBy(ctx context.Context, field, val string) (*User, error) {
	userQuery := fmt.Sprintf(`
		SELECT
			id,
			username,
			token,
			updated_at,
			created_at
		FROM
			"users"
		WHERE
			users.%s = $1
		LIMIT 1;
	`, field)

	user := User{}
	err := sqldb.QueryRow(ctx, userQuery, val).Scan(&user.ID, &user.Username, &user.Token, &user.UpdatedAt, &user.CreatedAt)

	if err != nil {
		log.Errorf("Could not query user for %s %v, %v", field, val, err)
		return nil, err
	}

	return &user, nil
}

func (user *User) Save(ctx context.Context) error {
	userQuery := `
		INSERT INTO "users" (username, token)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT user_username_unique DO UPDATE SET token = $2
		RETURNING id, updated_at, created_at;
	`

	err := sqldb.QueryRow(ctx, userQuery, user.Username, user.Token).Scan(&user.ID, &user.UpdatedAt, &user.CreatedAt)

	if err != nil {
		log.Errorf("Could not insert or update user, %v", err)
		return err
	}

	return nil
}
