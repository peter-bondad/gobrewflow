package inventory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryRepository interface {
	InsertInitialInventory(
		ctx context.Context,
		db bun.IDB,
		productID uuid.UUID,
	) error
	ChangeStock(
		ctx context.Context,
		db bun.IDB,
		params StockParams,
	) (*Inventory, error)
	FindByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryResult, error)
	FindInventoryByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error)
	SetStock(
		ctx context.Context,
		params SetStockParams,
	) (*Inventory, error)
}

type inventoryRepository struct {
	db bun.IDB
}

func NewInventoryRepository(db bun.IDB) InventoryRepository {
	return &inventoryRepository{
		db: db,
	}
}

type StockChange int

const (
	StockIncrease StockChange = 1
	StockDecrease StockChange = -1
)

type StockParams struct {
	ProductID uuid.UUID
	Quantity  int
	Change    StockChange
}

func (r *inventoryRepository) ChangeStock(
	ctx context.Context,
	db bun.IDB,
	params StockParams,
) (*Inventory, error) {
	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	if params.Change != StockIncrease && params.Change != StockDecrease {
		return nil, ErrInvalidStockChange
	}

	inventory := new(Inventory)

	err := db.NewSelect().
		Model(inventory).
		Where("product_id = ?", params.ProductID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	update := db.NewUpdate().
		Model(inventory).
		Where("product_id = ?", params.ProductID)

	if params.Change == StockIncrease {
		update.Set("quantity = quantity + ?", params.Quantity)
	} else {
		update.
			Where("quantity >= ?", params.Quantity).
			Set("quantity = quantity - ?", params.Quantity)
	}

	err = update.
		Set("updated_at = NOW()").
		Returning("*").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if params.Change == StockDecrease {
				return nil, ErrInsufficientStock
			}

			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	return inventory, nil
}
func (r *inventoryRepository) InsertInitialInventory(
	ctx context.Context,
	db bun.IDB,
	productID uuid.UUID,
) error {
	inventory := &Inventory{
		ProductID: productID,
	}

	_, err := db.NewInsert().
		Model(inventory).
		Exec(ctx)

	return err
}

type ProductInventoryResult struct {
	bun.BaseModel `bun:"table:inventory,alias:i"`
	ID            uuid.UUID
	ProductID     uuid.UUID
	Quantity      int
	ProductName   string
	SKU           string
	Price         int64
}

// query inventory with respective product
func (r *inventoryRepository) FindByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*ProductInventoryResult, error) {
	inventory := new(ProductInventoryResult)

	err := r.db.NewSelect().
		Model(inventory).
		Join("JOIN products AS p ON p.id = i.product_id").
		ColumnExpr("i.id").
		ColumnExpr("i.product_id").
		ColumnExpr("i.quantity").
		ColumnExpr("p.name AS product_name").
		ColumnExpr("p.sku").
		ColumnExpr("p.price").
		Where("i.product_id = ?", productID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	return inventory, nil
}

func (r *inventoryRepository) FindInventoryByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*Inventory, error) {
	inventory := new(Inventory)

	err := r.db.NewSelect().
		Model(inventory).
		Where("product_id = ?", productID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	return inventory, nil
}

type SetStockParams struct {
	ProductID uuid.UUID
	Quantity  int
}

func (r *inventoryRepository) SetStock(
	ctx context.Context,
	params SetStockParams,
) (*Inventory, error) {
	if params.Quantity < 0 {
		return nil, ErrInvalidQuantity
	}

	inventory := new(Inventory)

	err := r.db.NewUpdate().
		Model(inventory).
		Set("quantity = ?", params.Quantity).
		Set("updated_at = NOW()").
		Where("product_id = ?", params.ProductID).
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
