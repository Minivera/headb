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

type Permissions struct {
	ID         int64 `sql:"primary_key"`
	KeyID      int64
	DatabaseID *int64
	Role       Role
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
