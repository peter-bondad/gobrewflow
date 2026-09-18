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
}

type inventoryService struct {
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
