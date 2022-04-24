//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"encore.dev/types/uuid"
	"time"
)

type Users struct {
	ID        uuid.UUID `sql:"primary_key"`
	Username  *string
	Token     *string
	CreatedAt time.Time
	UpdatedAt time.Time
	UniqueID  *string
	Status    UserStatus
}
