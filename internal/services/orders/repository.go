package orders

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OrderRepository interface {
	InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error
	FindByID(ctx context.Context, db bun.IDB, id uuid.UUID) (*Orders, error)
}

type ordersRepository struct {
	db bun.IDB
}

func NewOrderRepository(db bun.IDB) OrderRepository {
	return &ordersRepository{db: db}
}

func (r *ordersRepository) InsertOrder(ctx context.Context, db bun.IDB, item *Orders) error {
	_, err := db.NewInsert().Model(item).Exec(ctx)
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
