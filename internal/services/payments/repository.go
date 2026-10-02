package payments

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PaymentRepository interface {
	Create(ctx context.Context, db bun.IDB, payment *Payment) error
	FindByID(ctx context.Context, db bun.IDB, id uuid.UUID) (*Payment, error)
	FindByOrderID(ctx context.Context, db bun.IDB, orderID uuid.UUID) (*Payment, error)
	FindByProviderCheckoutID(ctx context.Context, db bun.IDB, checkoutID string) (*Payment, error)
	Update(ctx context.Context, db bun.IDB, payment *Payment) error
}

type paymentRepository struct{}

func NewRepository() PaymentRepository {
	return &paymentRepository{}
}

func (r *paymentRepository) Create(
	ctx context.Context,
	db bun.IDB,
	payment *Payment,
) error {
	_, err := db.NewInsert().
		Model(payment).
		Exec(ctx)

	return err
}

func (r *paymentRepository) FindByID(
	ctx context.Context,
	db bun.IDB,
	id uuid.UUID,
) (*Payment, error) {
	payment := new(Payment)

	err := db.NewSelect().
		Model(payment).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *paymentRepository) FindByOrderID(
	ctx context.Context,
	db bun.IDB,
	orderID uuid.UUID,
) (*Payment, error) {
	payment := new(Payment)

	err := db.NewSelect().
		Model(payment).
		Where("order_id = ?", orderID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *paymentRepository) FindByProviderCheckoutID(
	ctx context.Context,
	db bun.IDB,
	checkoutID string,
) (*Payment, error) {
	payment := new(Payment)

	err := db.NewSelect().
		Model(payment).
		Where("provider_checkout_id = ?", checkoutID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *paymentRepository) Update(
	ctx context.Context,
	db bun.IDB,
	payment *Payment,
) error {
	_, err := db.NewUpdate().
		Model(payment).
		WherePK().
		Exec(ctx)

	return err
}
