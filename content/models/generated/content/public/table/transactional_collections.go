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

var TransactionalCollections = newTransactionalCollectionsTable("public", "transactional_collections", "")

type transactionalCollectionsTable struct {
	postgres.Table

	//Columns
	ID            postgres.ColumnString
	Name          postgres.ColumnString
	DatabaseID    postgres.ColumnString
	TransactionID postgres.ColumnString
	CreatedAt     postgres.ColumnTimestampz
	UpdatedAt     postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TransactionalCollectionsTable struct {
	transactionalCollectionsTable

	EXCLUDED transactionalCollectionsTable
}

// AS creates new TransactionalCollectionsTable with assigned alias
func (a TransactionalCollectionsTable) AS(alias string) *TransactionalCollectionsTable {
	return newTransactionalCollectionsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TransactionalCollectionsTable with assigned schema name
func (a TransactionalCollectionsTable) FromSchema(schemaName string) *TransactionalCollectionsTable {
	return newTransactionalCollectionsTable(schemaName, a.TableName(), a.Alias())
}

func newTransactionalCollectionsTable(schemaName, tableName, alias string) *TransactionalCollectionsTable {
	return &TransactionalCollectionsTable{
		transactionalCollectionsTable: newTransactionalCollectionsTableImpl(schemaName, tableName, alias),
		EXCLUDED:                      newTransactionalCollectionsTableImpl("", "excluded", ""),
	}
}

func newTransactionalCollectionsTableImpl(schemaName, tableName, alias string) transactionalCollectionsTable {
	var (
		IDColumn            = postgres.StringColumn("id")
		NameColumn          = postgres.StringColumn("name")
		DatabaseIDColumn    = postgres.StringColumn("database_id")
		TransactionIDColumn = postgres.StringColumn("transaction_id")
		CreatedAtColumn     = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn     = postgres.TimestampzColumn("updated_at")
		allColumns          = postgres.ColumnList{IDColumn, NameColumn, DatabaseIDColumn, TransactionIDColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns      = postgres.ColumnList{NameColumn, DatabaseIDColumn, TransactionIDColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return transactionalCollectionsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:            IDColumn,
		Name:          NameColumn,
		DatabaseID:    DatabaseIDColumn,
		TransactionID: TransactionIDColumn,
		CreatedAt:     CreatedAtColumn,
		UpdatedAt:     UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
