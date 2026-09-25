package inventory_movements

import (
	"context"
	"time"

	"gobrewflow/internal/services/inventory"
	"gobrewflow/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryMovementsService interface {
	RecordMovement(
		ctx context.Context,
		tx bun.IDB,
		input inventory.RecordMovementInput,
	) error

	ListMovements(ctx context.Context, input ListMovementsInput) (ListMovementsOutput, error)
}

type inventoryMovementsService struct {
	movementsRepo InventoryMovementsRepository
}

func NewInventoryMovementsService(
	movementsRepo InventoryMovementsRepository,
) *inventoryMovementsService {
	return &inventoryMovementsService{
		movementsRepo: movementsRepo,
	}
}

func (s *inventoryMovementsService) RecordMovement(
	ctx context.Context,
	tx bun.IDB,
	input inventory.RecordMovementInput,
) error {
	if input.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	if !isValidMovementType(input.Type) {
		return ErrInvalidMovementType
	}

	movement := &InventoryMovement{
		ID:          uuid.New(),
		ProductID:   input.ProductID,
		Type:        input.Type,
		Quantity:    input.Quantity,
		BeforeStock: input.BeforeStock,
		AfterStock:  input.AfterStock,
		CreatedAt:   time.Now(),
	}

	return s.movementsRepo.CreateMovement(ctx, tx, movement)
}

type InventoryMovementOutput struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Type      inventory.MovementType
	Quantity  int
	CreatedAt time.Time
}

func isValidMovementType(t inventory.MovementType) bool {
	switch t {
	case inventory.MovementTypeReceived, inventory.MovementTypeSold,
		inventory.MovementTypeReturned, inventory.MovementTypeDamaged,
		inventory.MovementTypeAdjusted:
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
