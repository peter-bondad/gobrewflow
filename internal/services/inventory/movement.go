package inventory

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MovementType string

const (
	MovementTypeReceived MovementType = "RECEIVED"
	MovementTypeReturned MovementType = "RETURNED"
	MovementTypeDamaged  MovementType = "DAMAGED"
	MovementTypeSold     MovementType = "SOLD"
	MovementTypeAdjusted MovementType = "ADJUSTED"
)

type RecordMovementInput struct {
	ProductID   uuid.UUID
	Type        MovementType
	Quantity    int
	BeforeStock int
	AfterStock  int
}

type MovementRecorder interface {
	RecordMovement(
		ctx context.Context,
		tx bun.IDB,
		input RecordMovementInput,
	) error
}
