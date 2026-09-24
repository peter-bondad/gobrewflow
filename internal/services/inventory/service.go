package inventory

import (
	"context"
	"database/sql"
	"errors"
	"gobrewflow/internal/database"

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
		adjustedStock int,
	) (*AdjustStockOutput, error)
}

type MovementRecorder interface {
	RecordAdjustment(ctx context.Context, tx bun.IDB, productID uuid.UUID, beforeStock, afterStock int) error
}

type inventoryService struct {
	inventoryRepo    InventoryRepository
	movementRecorder MovementRecorder
	txManager        database.TxManager
}

func NewInventoryService(inventoryRepo InventoryRepository, movementRecorder MovementRecorder, txManager database.TxManager) InventoryService {
	return &inventoryService{
		inventoryRepo:    inventoryRepo,
		movementRecorder: movementRecorder,
		txManager:        txManager,
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

type AdjustStockOutput struct {
	ProductID   uuid.UUID
	BeforeStock int
	AfterStock  int
}

func (s *inventoryService) AdjustStock(
	ctx context.Context,
	productID uuid.UUID,
	adjustedStock int,
) (*AdjustStockOutput, error) {
	if adjustedStock < 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	beforeStock := inventory.Quantity

	err = s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		_, err := s.inventoryRepo.SetStock(
			ctx,
			tx,
			SetStockParams{
				ProductID: productID,
				Quantity:  adjustedStock,
			},
		)
		if err != nil {
			return err
		}

		return s.movementRecorder.RecordAdjustment(
			ctx,
			tx,
			productID,
			beforeStock,
			adjustedStock,
		)
	})

	if err != nil {
		return nil, err
	}

	return &AdjustStockOutput{
		ProductID:   productID,
		BeforeStock: beforeStock,
		AfterStock:  adjustedStock,
	}, nil
}
