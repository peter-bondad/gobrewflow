package inventory_movements

import (
	"context"
	"time"

	"gobrewflow/internal/database"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryMovementsService interface {
	CreateMovement(ctx context.Context, input CreateMovementInput) (*InventoryMovementOutput, error)
	ListMovements(ctx context.Context, input ListMovementsInput) (ListMovementsOutput, error)
}

type inventoryMovementsService struct {
	movementsRepo InventoryMovementsRepository
	inventoryRepo inventory.InventoryRepository
	txManager     database.TxManager
}

func NewInventoryMovementsService(movementsRepo InventoryMovementsRepository, inventoryRepo inventory.InventoryRepository) InventoryMovementsService {
	return &inventoryMovementsService{
		movementsRepo: movementsRepo,
		inventoryRepo: inventoryRepo,
	}
}

type CreateMovementInput struct {
	ProductID uuid.UUID
	Type      InventoryMovementType
	Quantity  int
}

type InventoryMovementOutput struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Type      InventoryMovementType
	Quantity  int
	CreatedAt time.Time
}

func (s *inventoryMovementsService) CreateMovement(
	ctx context.Context,
	input CreateMovementInput,
) (*InventoryMovementOutput, error) {
	if input.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	if !isValidMovementType(input.Type) {
		return nil, ErrInvalidMovementType
	}

	movement := &InventoryMovement{
		ID:        uuid.New(),
		ProductID: input.ProductID,
		Type:      input.Type,
		Quantity:  input.Quantity,
		CreatedAt: time.Now(),
	}

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		switch input.Type {
		case InventoryMovementTypeReceived,
			InventoryMovementTypeReturned:

			_, err := s.inventoryRepo.ChangeStock(ctx, tx, inventory.StockParams{
				ProductID: input.ProductID,
				Quantity:  input.Quantity,
				Change:    inventory.StockIncrease,
			})
			if err != nil {
				return err
			}

		case InventoryMovementTypeSold,
			InventoryMovementTypeDamaged:

			_, err := s.inventoryRepo.ChangeStock(ctx, tx, inventory.StockParams{
				ProductID: input.ProductID,
				Quantity:  input.Quantity,
				Change:    inventory.StockDecrease,
			})
			if err != nil {
				return err
			}

		case InventoryMovementTypeAdjusted:
			// Define adjustment behavior separately.
			// Currenly planning
		}

		return s.movementsRepo.CreateMovement(ctx, tx, movement)
	})

	if err != nil {
		return nil, err
	}

	return &InventoryMovementOutput{
		ID:        movement.ID,
		ProductID: movement.ProductID,
		Type:      movement.Type,
		Quantity:  movement.Quantity,
		CreatedAt: movement.CreatedAt,
	}, nil
}

func isValidMovementType(t InventoryMovementType) bool {
	switch t {
	case InventoryMovementTypeReceived, InventoryMovementTypeSold,
		InventoryMovementTypeReturned, InventoryMovementTypeDamaged,
		InventoryMovementTypeAdjusted:
		return true
	}
	return false
}

type ListMovementsInput struct {
	ProductID *uuid.UUID
	Page      int
	Limit     int
}

type ListMovementsOutput struct {
	Data       []InventoryMovementOutput
	Pagination shared.Pagination
}

func (s *inventoryMovementsService) ListMovements(ctx context.Context, input ListMovementsInput) (ListMovementsOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Page <= 0 {
		input.Page = 1
	}

	offset := (input.Page - 1) * input.Limit

	params := ListMovementsParams{
		ProductID: input.ProductID,
		Limit:     input.Limit,
		Offset:    offset,
	}

	movements, err := s.movementsRepo.ListMovements(ctx, params)
	if err != nil {
		return ListMovementsOutput{}, err
	}

	total, err := s.movementsRepo.CountMovements(ctx, params)
	if err != nil {
		return ListMovementsOutput{}, err
	}

	outputs := make([]InventoryMovementOutput, len(movements))
	for i, m := range movements {
		outputs[i] = InventoryMovementOutput{
			ID:        m.ID,
			ProductID: m.ProductID,
			Type:      m.Type,
			Quantity:  m.Quantity,
			CreatedAt: m.CreatedAt,
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + input.Limit - 1) / input.Limit
	}

	return ListMovementsOutput{
		Data: outputs,
		Pagination: shared.Pagination{
			Page:       input.Page,
			Limit:      input.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
