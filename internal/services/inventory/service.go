package inventory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryService interface {
	CreateInitialInventory(
		ctx context.Context,
		tx bun.IDB,
		productID uuid.UUID,
	) error
	FindByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryOutput, error)
	GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryQuantityResponse, error)
	AdjustStock(ctx context.Context, productID uuid.UUID, input AdjustStockInput) (AdjustStockOutput, error)
}

type inventoryService struct {
	db            bun.IDB
	inventoryRepo InventoryRepository
}

func NewInventoryService(inventoryRepo InventoryRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
	}
}

type ProductInventoryOutput struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	Quantity    int
	ProductName string
	SKU         string
	Price       int64
}

func (s *inventoryService) CreateInitialInventory(
	ctx context.Context,
	tx bun.IDB,
	productID uuid.UUID,
) error {
	return s.inventoryRepo.InsertInitialInventory(ctx, tx, productID)
}

func (s inventoryService) FindByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryOutput, error) {

	inventoryProduct, err := s.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return &ProductInventoryOutput{
		ID:          inventoryProduct.ID,
		ProductID:   inventoryProduct.ProductID,
		ProductName: inventoryProduct.ProductName,
		Quantity:    inventoryProduct.Quantity,
		Price:       inventoryProduct.Price,
		SKU:         inventoryProduct.SKU,
	}, nil
}

func (s inventoryService) GetInventoryByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryQuantityResponse, error) {
	inventory, err := s.inventoryRepo.FindInventoryByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return &ProductInventoryQuantityResponse{
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
	}, nil
}

type AdjustStockInput struct {
	Quantity int         `json:"quantity"`
	Change   StockChange `json:"change"`
}

type AdjustStockOutput struct {
	ProductID   uuid.UUID `json:"product_id"`
	BeforeStock int       `json:"before_stock"`
	Adjustment  int       `json:"adjustment"`
	AfterStock  int       `json:"after_stock"`
}

func (s inventoryService) AdjustStock(
	ctx context.Context,
	productID uuid.UUID,
	input AdjustStockInput,
) (AdjustStockOutput, error) {

	inventory, err := s.inventoryRepo.ChangeStock(
		ctx,
		s.db,
		StockParams{
			ProductID: productID,
			Quantity:  input.Quantity,
			Change:    input.Change,
		},
	)
	if err != nil {
		return AdjustStockOutput{}, err
	}

	adjustment := input.Quantity

	switch input.Change {
	case StockIncrease:
		// Quantity stays positive.
		adjustment = +input.Quantity
	case StockDecrease:
		// Represent deduction as a negative adjustment.
		adjustment = -input.Quantity
	default:
		return AdjustStockOutput{}, ErrInvalidStockChange
	}

	return AdjustStockOutput{
		ProductID:   productID,
		BeforeStock: inventory.Quantity - adjustment,
		Adjustment:  adjustment,
		AfterStock:  inventory.Quantity,
	}, nil
}
