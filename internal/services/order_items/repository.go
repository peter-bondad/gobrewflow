package order_items

import (
	"context"

	"github.com/uptrace/bun"
)

type OrderItemRepository interface {
	InsertOrderItem(ctx context.Context, item *OrderItem) error
}

type orderItemRepository struct {
	db bun.IDB
}

func NewOrderItemRepository(db bun.IDB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) InsertOrderItem(ctx context.Context, item *OrderItem) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	return err
}
