package dbhandler

import (
	"context"
	"log"

	"database/sql"

	"github.com/behrang/sqlbatch"
	"github.com/lib/pq"
)

// DBHandler contains a connection to database.
type DBHandler struct {
	DB *sql.DB
}

// Batch creates a transaction and executes the batch of commands in that transaction.
// If a retryable error is received, the batch is retried.
func (handler DBHandler) Batch(opts *sql.TxOptions, commands []sqlbatch.Command) ([]interface{}, error) {

	for {
		results, err := handler.tryBatch(opts, commands)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
			log.Printf("🟡 Retryable Postgres error, retrying: %v", err)
			continue
		}
		return results, err
	}
}

func (handler DBHandler) tryBatch(opts *sql.TxOptions, commands []sqlbatch.Command) (results []interface{}, err error) {

	results = make([]interface{}, len(commands))

	tx, err := handler.DB.BeginTx(context.Background(), opts)
	if err != nil {
		return
	}
	defer tx.Rollback()

	results, err = sqlbatch.Batch(tx, commands)

	if err == nil {
		err = tx.Commit()
	}

	return
}
