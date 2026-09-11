package inventory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryRepositoryInterface interface {
	AddStock(ctx context.Context, params StockParams) (*Inventory, error)
	RemoveStock(ctx context.Context, params StockParams) (*Inventory, error)
}

type inventoryRepository struct {
	db bun.IDB
}

func NewInventoryRepository(db bun.IDB) InventoryRepositoryInterface {
	return &inventoryRepository{
		db: db,
	}
}

type StockParams struct {
	ProductID uuid.UUID
	Quantity  int
}

func (r *inventoryRepository) AddStock(ctx context.Context, params StockParams) (*Inventory, error) {
	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory := new(Inventory)

	// Check first if the inventory exists
	err := r.db.NewSelect().
		Model(inventory).
		Where("product_id = ?", params.ProductID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	err = r.db.NewUpdate().
		Model(inventory).
		Where("product_id = ?", params.ProductID).
		Set("quantity = quantity + ?", params.Quantity).
		Set("updated_at = NOW()").
		Returning("*").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	return inventory, nil
}

func (r *inventoryRepository) RemoveStock(ctx context.Context, params StockParams) (*Inventory, error) {
	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory := new(Inventory)

	// Check first if the inventory exists.
	err := r.db.NewSelect().
		Model(inventory).
		Where("product_id = ?", params.ProductID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	// if exists
	// remove/decrease the quantity if the stored quantity is greater than the inputted quantity
	err = r.db.NewUpdate().
		Model(inventory).
		Where("product_id = ?", params.ProductID).
		Where("quantity >= ?", params.Quantity).
		Set("quantity = quantity - ?", params.Quantity).
		Set("updated_at = NOW()").
		Returning("*").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Could be either:
			// 1. inventory doesn't exist
			// 2. not enough stock
			return nil, ErrInsufficientStock
		}

		return nil, err
	}

	return inventory, nil
}
