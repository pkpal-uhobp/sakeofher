package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"sakeofher/internal/repository/pool"
)

type Manager struct {
	pool      *pool.ConnectionPool
	opTimeout time.Duration
}

func NewManager(db *pool.ConnectionPool) *Manager {
	return &Manager{pool: db, opTimeout: db.QueryTimeout()}
}

func (m *Manager) Querier(ctx context.Context) Querier {
	if q, ok := FromContext(ctx); ok {
		return q
	}
	return m.pool.Pool
}

func (m *Manager) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := FromContext(ctx); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, m.opTimeout)
}

func (m *Manager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.WithinTransactionOptions(ctx, pgx.TxOptions{}, fn)
}

func (m *Manager) WithinTransactionOptions(ctx context.Context, opts pgx.TxOptions, fn func(ctx context.Context) error) error {
	if _, ok := FromContext(ctx); ok {
		return fn(ctx)
	}

	txCtx, cancel := context.WithTimeout(ctx, m.opTimeout)
	defer cancel()

	dbTx, err := m.pool.BeginTx(txCtx, opts)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx = injectTx(txCtx, dbTx)

	defer func() {
		_ = dbTx.Rollback(context.Background())
	}()

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := dbTx.Commit(txCtx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
