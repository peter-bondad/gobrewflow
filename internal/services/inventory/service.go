package inventory

import (
	"context"
	"database/sql"
	"errors"
	"gobrewflow/internal/services/products"

	"github.com/google/uuid"
)

type InventoryServiceInterface interface {
	FindByProductID(ctx context.Context, productID uuid.UUID) (*ProductInventoryOutput, error)
}

type inventoryService struct {
	inventoryRepo InventoryRepositoryInterface
	productRepo   products.ProductRepositoryInterface
}

func NewInventoryService(inventoryRepo InventoryRepositoryInterface, productRepo products.ProductRepositoryInterface) InventoryServiceInterface {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
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
