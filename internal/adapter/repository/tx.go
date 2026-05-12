package repository

import (
	"context"
	"fmt"

	"be-modami-user-service/pkg/pgxtx"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX aliases pgxtx.DBTX so repo files don't need an extra import.
type DBTX = pgxtx.DBTX

// TxManager propagates a pgx transaction through context so all repo calls in fn share it.
type TxManager struct{ db *pgxpool.Pool }

func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	ctx = pgxtx.Inject(ctx, tx)
	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// dbFromCtx returns the pgx.Tx from ctx if present, otherwise the pool.
func dbFromCtx(ctx context.Context, db *pgxpool.Pool) DBTX {
	return pgxtx.Querier(ctx, db)
}
