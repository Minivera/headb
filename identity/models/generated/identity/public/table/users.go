//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Users = newUsersTable("public", "users", "")

type usersTable struct {
	postgres.Table

	//Columns
	ID        postgres.ColumnInteger
	Username  postgres.ColumnString
	Token     postgres.ColumnString
	CreatedAt postgres.ColumnTimestampz
	UpdatedAt postgres.ColumnTimestampz
	UniqueID  postgres.ColumnString
	Status    postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type UsersTable struct {
	usersTable

	EXCLUDED usersTable
}

// AS creates new UsersTable with assigned alias
func (a UsersTable) AS(alias string) *UsersTable {
	return newUsersTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UsersTable with assigned schema name
func (a UsersTable) FromSchema(schemaName string) *UsersTable {
	return newUsersTable(schemaName, a.TableName(), a.Alias())
}

func newUsersTable(schemaName, tableName, alias string) *UsersTable {
	return &UsersTable{
		usersTable: newUsersTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newUsersTableImpl("", "excluded", ""),
	}
}

func newUsersTableImpl(schemaName, tableName, alias string) usersTable {
	var (
		IDColumn        = postgres.IntegerColumn("id")
		UsernameColumn  = postgres.StringColumn("username")
		TokenColumn     = postgres.StringColumn("token")
		CreatedAtColumn = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn = postgres.TimestampzColumn("updated_at")
		UniqueIDColumn  = postgres.StringColumn("unique_id")
		StatusColumn    = postgres.StringColumn("status")
		allColumns      = postgres.ColumnList{IDColumn, UsernameColumn, TokenColumn, CreatedAtColumn, UpdatedAtColumn, UniqueIDColumn, StatusColumn}
		mutableColumns  = postgres.ColumnList{UsernameColumn, TokenColumn, CreatedAtColumn, UpdatedAtColumn, UniqueIDColumn, StatusColumn}
	)

	return usersTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		Username:  UsernameColumn,
		Token:     TokenColumn,
		CreatedAt: CreatedAtColumn,
		UpdatedAt: UpdatedAtColumn,
		UniqueID:  UniqueIDColumn,
		Status:    StatusColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
