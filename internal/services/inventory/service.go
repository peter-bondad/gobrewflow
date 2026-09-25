package inventory

import (
	"context"
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

	FindByProductID(
		ctx context.Context,
		productID uuid.UUID,
	) (*ProductInventoryOutput, error)

	GetInventoryByProductID(
		ctx context.Context,
		productID uuid.UUID,
	) (*ProductInventoryQuantityResponse, error)

	AdjustStock(
		ctx context.Context,
		productID uuid.UUID,
		adjustedStock int,
	) (*AdjustStockOutput, error)

	ReceiveStock(
		ctx context.Context,
		productID uuid.UUID,
		receivedStock int,
	) (*ReceiveStockOutput, error)

	ReturnStock(
		ctx context.Context,
		productID uuid.UUID,
		returnStock int,
	) (*ReturnStockOutput, error)

	DamageStock(
		ctx context.Context,
		productID uuid.UUID,
		damagedStock int,
	) (*DamageStockOutput, error)
}

type inventoryService struct {
	inventoryRepo    InventoryRepository
	movementRecorder MovementRecorder
	txManager        database.TxManager
}

func NewInventoryService(
	inventoryRepo InventoryRepository,
	movementRecorder MovementRecorder,
	txManager database.TxManager,
) InventoryService {
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

func (s *inventoryService) FindByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*ProductInventoryOutput, error) {
	inventoryProduct, err := s.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
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

func (s *inventoryService) GetInventoryByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*ProductInventoryQuantityResponse, error) {
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

	var output AdjustStockOutput

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		inventory, err := s.inventoryRepo.FindByProductID(ctx, productID)
		if err != nil {
			return err
		}

		beforeStock := inventory.Quantity

		// Nothing actually changed.
		if beforeStock == adjustedStock {
			output = AdjustStockOutput{
				ProductID:   productID,
				BeforeStock: beforeStock,
				AfterStock:  adjustedStock,
			}
			return nil
		}

		_, err = s.inventoryRepo.SetStock(
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

		quantity := adjustedStock - beforeStock
		if quantity < 0 {
			quantity = -quantity
		}

		err = s.movementRecorder.RecordMovement(
			ctx,
			tx,
			RecordMovementInput{
				ProductID:   productID,
				Type:        MovementTypeAdjusted,
				Quantity:    quantity,
				BeforeStock: beforeStock,
				AfterStock:  adjustedStock,
			},
		)
		if err != nil {
			return err
		}

		output = AdjustStockOutput{
			ProductID:   productID,
			BeforeStock: beforeStock,
			AfterStock:  adjustedStock,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &output, nil
}

type ReceiveStockOutput struct {
	ProductID   uuid.UUID
	BeforeStock int
	AfterStock  int
}

func (s *inventoryService) ReceiveStock(
	ctx context.Context,
	productID uuid.UUID,
	receivedStock int,
) (*ReceiveStockOutput, error) {
	if receivedStock <= 0 {
		return nil, ErrInvalidQuantity
	}

	var output ReceiveStockOutput

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		inventory, err := s.inventoryRepo.ChangeStock(
			ctx,
			tx,
			StockParams{
				ProductID: productID,
				Quantity:  receivedStock,
				Change:    StockIncrease,
			},
		)
		if err != nil {
			return err
		}

		afterStock := inventory.Quantity
		beforeStock := afterStock - receivedStock

		err = s.movementRecorder.RecordMovement(
			ctx,
			tx,
			RecordMovementInput{
				ProductID:   productID,
				Type:        MovementTypeReceived,
				Quantity:    receivedStock,
				BeforeStock: beforeStock,
				AfterStock:  afterStock,
			},
		)
		if err != nil {
			return err
		}

		output = ReceiveStockOutput{
			ProductID:   productID,
			BeforeStock: beforeStock,
			AfterStock:  afterStock,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &output, nil
}

type ReturnStockOutput struct {
	ProductID   uuid.UUID
	BeforeStock int
	AfterStock  int
}

func (s *inventoryService) ReturnStock(
	ctx context.Context,
	productID uuid.UUID,
	returnStock int,
) (*ReturnStockOutput, error) {
	if returnStock <= 0 {
		return nil, ErrInvalidQuantity
	}

	var output ReturnStockOutput

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		inventory, err := s.inventoryRepo.ChangeStock(
			ctx,
			tx,
			StockParams{
				ProductID: productID,
				Quantity:  returnStock,
				Change:    StockIncrease,
			},
		)
		if err != nil {
			return err
		}

		afterStock := inventory.Quantity
		beforeStock := afterStock - returnStock

		err = s.movementRecorder.RecordMovement(
			ctx,
			tx,
			RecordMovementInput{
				ProductID:   productID,
				Type:        MovementTypeReturned,
				Quantity:    returnStock,
				BeforeStock: beforeStock,
				AfterStock:  afterStock,
			},
		)
		if err != nil {
			return err
		}

		output = ReturnStockOutput{
			ProductID:   productID,
			BeforeStock: beforeStock,
			AfterStock:  afterStock,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &output, nil
}

type DamageStockOutput struct {
	ProductID   uuid.UUID
	BeforeStock int
	AfterStock  int
}

func (s *inventoryService) DamageStock(
	ctx context.Context,
	productID uuid.UUID,
	damagedStock int,
) (*DamageStockOutput, error) {
	if damagedStock <= 0 {
		return nil, ErrInvalidQuantity
	}

	var output DamageStockOutput

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		inventory, err := s.inventoryRepo.ChangeStock(
			ctx,
			tx,
			StockParams{
				ProductID: productID,
				Quantity:  damagedStock,
				Change:    StockDecrease,
			},
		)
		if err != nil {
			return err
		}

		afterStock := inventory.Quantity
		beforeStock := afterStock + damagedStock

		err = s.movementRecorder.RecordMovement(
			ctx,
			tx,
			RecordMovementInput{
				ProductID:   productID,
				Type:        MovementTypeDamaged,
				Quantity:    damagedStock,
				BeforeStock: beforeStock,
				AfterStock:  afterStock,
			},
		)
		if err != nil {
			return err
		}

		output = DamageStockOutput{
			ProductID:   productID,
			BeforeStock: beforeStock,
			AfterStock:  afterStock,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &output, nil
}
