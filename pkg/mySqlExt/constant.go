package mySqlExt

type ctxKey string

const (
	// CtxSqlTx is the context key for sql transaction
	CtxSqlTx ctxKey = "sql_tx"

	// CtxSQLTableNameKey is the context key for sql table name
	CtxSQLTableNameKey ctxKey = "table_name"
)
