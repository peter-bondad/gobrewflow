package database

import (
	"context"

	"github.com/uptrace/bun"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx bun.IDB) error) error
}

type txManager struct {
	db *bun.DB
}

func NewTxManager(db *bun.DB) TxManager {
	return &txManager{
		db: db,
	}
}

func (m *txManager) WithTx(
	ctx context.Context,
	fn func(tx bun.IDB) error,
) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
