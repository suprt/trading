package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type UoW struct {
	pool     *pgxpool.Pool
	reporter ErrorReporter
}
type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}

func NewUoW(pool *pgxpool.Pool, reporter ErrorReporter) *UoW {
	return &UoW{pool: pool, reporter: reporter}
}

func (u *UoW) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(context.Background()); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				u.reporter.Report(context.Background(), rbErr, map[string]string{
					"op": "uow_rollback_after_panic",
				})
			}
			panic(r) //пробрасываем панику дальше - panic interceptor поймает и логирует её
		}

		if rbErr := tx.Rollback(context.Background()); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			u.reporter.Report(context.Background(), rbErr, map[string]string{
				"op": "uow_rollback",
			})
		}
	}()

	ctx = WithTx(ctx, tx)

	if err := fn(ctx); err != nil {
		return err
	}
	//TODO: recover
	return tx.Commit(ctx)
}

type txKey struct{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func DBFromContext(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return pool
}
