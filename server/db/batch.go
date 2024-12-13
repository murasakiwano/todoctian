// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: batch.go

package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrBatchAlreadyClosed = errors.New("batch already closed")
)

const batchUpdateTaskOrders = `-- name: BatchUpdateTaskOrders :batchexec
UPDATE tasks
SET "order" = $2
WHERE tasks.id = $1
`

type BatchUpdateTaskOrdersBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type BatchUpdateTaskOrdersParams struct {
	ID    pgtype.UUID
	Order int32
}

func (q *Queries) BatchUpdateTaskOrders(ctx context.Context, arg []BatchUpdateTaskOrdersParams) *BatchUpdateTaskOrdersBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.ID,
			a.Order,
		}
		batch.Queue(batchUpdateTaskOrders, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &BatchUpdateTaskOrdersBatchResults{br, len(arg), false}
}

func (b *BatchUpdateTaskOrdersBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *BatchUpdateTaskOrdersBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}
