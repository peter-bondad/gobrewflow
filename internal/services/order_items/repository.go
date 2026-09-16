package order_items

import (
	"context"

	"github.com/uptrace/bun"
)

type OrderItemRepository interface {
	InsertOrderItem(ctx context.Context, item *OrderItem) error
}

type orderItemRepository struct {
	db *bun.DB
}

func NewOrderItemRepository(db *bun.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) InsertOrderItem(ctx context.Context, item *OrderItem) error {
	_, err := r.db.NewInsert().Model(item).Exec(ctx)
	return err
}
