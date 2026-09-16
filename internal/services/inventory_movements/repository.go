package inventory_movements

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryMovementsRepository interface {
	CreateMovement(ctx context.Context, movement *InventoryMovement) error
	ListMovements(ctx context.Context, params ListMovementsParams) ([]InventoryMovement, error)
	CountMovements(ctx context.Context, params ListMovementsParams) (int, error)
}

type inventoryMovementsRepository struct {
	db bun.IDB
}

func NewInventoryMovementsRepository(db bun.IDB) InventoryMovementsRepository {
	return &inventoryMovementsRepository{
		db: db,
	}
}

func (r *inventoryMovementsRepository) CreateMovement(ctx context.Context, movement *InventoryMovement) error {
	_, err := r.db.NewInsert().Model(movement).Exec(ctx)
	return err
}

func (r *inventoryMovementsRepository) FindByID(ctx context.Context, id uuid.UUID) (*InventoryMovement, error) {
	movement := new(InventoryMovement)
	err := r.db.NewSelect().Model(movement).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovementNotFound
		}
		return nil, err
	}
	return movement, nil
}

type ListMovementsParams struct {
	ProductID *uuid.UUID
	Page      int
	Limit     int
	Offset    int
}

func (r *inventoryMovementsRepository) ListMovements(ctx context.Context, params ListMovementsParams) ([]InventoryMovement, error) {
	movements := make([]InventoryMovement, 0)

	query := r.db.NewSelect().Model(&movements)

	if params.ProductID != nil {
		query = query.Where("product_id = ?", *params.ProductID)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	err := query.Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return movements, nil
}

func (r *inventoryMovementsRepository) CountMovements(ctx context.Context, params ListMovementsParams) (int, error) {
	query := r.db.NewSelect().Model((*InventoryMovement)(nil))

	if params.ProductID != nil {
		query = query.Where("product_id = ?", *params.ProductID)
	}

	return query.Count(ctx)
}
