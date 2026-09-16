package orders

import (
	"context"

	"github.com/uptrace/bun"
)

type OrderRepository interface {
	InsertOrderItem(ctx context.Context, item *Orders) error
}

type ordersRepository struct {
	db bun.IDB
}

func NewOrderItemRepository(db bun.IDB) OrderRepository {
	return &ordersRepository{db: db}
}

func (r *ordersRepository) InsertOrderItem(ctx context.Context, item *Orders) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	return err
}
