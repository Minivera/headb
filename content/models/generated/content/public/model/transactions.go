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

type Transactions struct {
	ID         uuid.UUID `sql:"primary_key"`
	Script     string
	DatabaseID uuid.UUID
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}