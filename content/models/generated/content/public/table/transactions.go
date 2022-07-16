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

var Transactions = newTransactionsTable("public", "transactions", "")

type transactionsTable struct {
	postgres.Table

	//Columns
	ID         postgres.ColumnString
	Script     postgres.ColumnString
	DatabaseID postgres.ColumnString
	ExpiresAt  postgres.ColumnTimestampz
	CreatedAt  postgres.ColumnTimestampz
	UpdatedAt  postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TransactionsTable struct {
	transactionsTable

	EXCLUDED transactionsTable
}

// AS creates new TransactionsTable with assigned alias
func (a TransactionsTable) AS(alias string) *TransactionsTable {
	return newTransactionsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TransactionsTable with assigned schema name
func (a TransactionsTable) FromSchema(schemaName string) *TransactionsTable {
	return newTransactionsTable(schemaName, a.TableName(), a.Alias())
}

func newTransactionsTable(schemaName, tableName, alias string) *TransactionsTable {
	return &TransactionsTable{
		transactionsTable: newTransactionsTableImpl(schemaName, tableName, alias),
		EXCLUDED:          newTransactionsTableImpl("", "excluded", ""),
	}
}

func newTransactionsTableImpl(schemaName, tableName, alias string) transactionsTable {
	var (
		IDColumn         = postgres.StringColumn("id")
		ScriptColumn     = postgres.StringColumn("script")
		DatabaseIDColumn = postgres.StringColumn("database_id")
		ExpiresAtColumn  = postgres.TimestampzColumn("expires_at")
		CreatedAtColumn  = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn  = postgres.TimestampzColumn("updated_at")
		allColumns       = postgres.ColumnList{IDColumn, ScriptColumn, DatabaseIDColumn, ExpiresAtColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns   = postgres.ColumnList{ScriptColumn, DatabaseIDColumn, ExpiresAtColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return transactionsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:         IDColumn,
		Script:     ScriptColumn,
		DatabaseID: DatabaseIDColumn,
		ExpiresAt:  ExpiresAtColumn,
		CreatedAt:  CreatedAtColumn,
		UpdatedAt:  UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
