//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type APIKeys struct {
	ID         int64 `sql:"primary_key"`
	Value      string
	UserID     int64
	LastUsedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
