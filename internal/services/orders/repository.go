package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OrderRepository interface {
	InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error
	FindByID(ctx context.Context, db bun.IDB, id uuid.UUID) (*Orders, error)
	UpdateOrderStatus(ctx context.Context, db bun.IDB, orderID uuid.UUID, status OrderStatus) error
}

type ordersRepository struct {
	db bun.IDB
}

func NewOrderRepository(db bun.IDB) OrderRepository {
	return &ordersRepository{db: db}
}

func (r *ordersRepository) InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error {
	_, err := db.NewInsert().Model(item).Returning("*").Exec(ctx)
	return err
}

func (r *ordersRepository) FindByID(
	ctx context.Context,
	db bun.IDB,
	id uuid.UUID,
) (*Orders, error) {
	order := new(Orders)

	err := db.NewSelect().
		Model(order).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *ordersRepository) UpdateOrderStatus(
	ctx context.Context,
	db bun.IDB,
	orderID uuid.UUID,
	status OrderStatus,
) error {
	_, err := db.NewUpdate().
		Model((*Orders)(nil)).
		Set("status = ?", status).
		Where("id = ?", orderID).
		Exec(ctx)

	return err
}
