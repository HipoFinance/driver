package repository

import (
	"database/sql"

	"github.com/behrang/sqlbatch"
)

var (
	BatchOptionNormal = sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelReadCommitted,
	}

	BatchOptionNormalReadOnly = sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelReadCommitted,
	}

	BatchOptionSerializable = sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelSerializable,
	}
)

// BatchHandler is a database handler that executes a batch of SQL commands.
type BatchHandler interface {
	Batch(opts *sql.TxOptions, commands []sqlbatch.Command) ([]interface{}, error)
}
