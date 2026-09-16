package orders

import (
	"context"

	"github.com/uptrace/bun"
)

type OrderRepository interface {
	BeginTx(ctx context.Context) (bun.Tx, error)
	InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error
}

type ordersRepository struct {
	db bun.IDB
}

func NewOrderItemRepository(db bun.IDB) OrderRepository {
	return &ordersRepository{db: db}
}

func (r *ordersRepository) BeginTx(ctx context.Context) (bun.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *ordersRepository) InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error {
	_, err := db.NewInsert().Model(item).Exec(ctx)
	return err
}
