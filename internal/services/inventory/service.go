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
	AdjustStock(
		ctx context.Context,
		productID uuid.UUID,
		input AdjustStockInput,
	) (*AdjustStockOutput, error)
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
	Quantity int
}

type AdjustStockOutput struct {
	ProductID   uuid.UUID `json:"productId"`
	BeforeStock int       `json:"beforeStock"`
	Adjustment  int       `json:"adjustment"`
	AfterStock  int       `json:"afterStock"`
}

func (s *inventoryService) AdjustStock(
	ctx context.Context,
	productID uuid.UUID,
	input AdjustStockInput,
) (*AdjustStockOutput, error) {
	if input.Quantity < 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	beforeStock := inventory.Quantity
	afterStock := input.Quantity
	adjustment := afterStock - beforeStock

	_, err = s.inventoryRepo.SetStock(
		ctx,
		s.db,
		SetStockParams{
			ProductID: productID,
			Quantity:  afterStock,
		},
	)
	if err != nil {
		return nil, err
	}

	return &AdjustStockOutput{
		ProductID:   productID,
		BeforeStock: beforeStock,
		Adjustment:  adjustment,
		AfterStock:  afterStock,
	}, nil
}
