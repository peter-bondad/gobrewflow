package inventory_movements

import (
	"gobrewflow/internal/services/inventory"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type InventoryMovement struct {
	bun.BaseModel `bun:"table:inventory_movements,alias:im"`

	ID          uuid.UUID              `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ProductID   uuid.UUID              `bun:"product_id,notnull,type:uuid"`
	Type        inventory.MovementType `bun:"type,notnull"`
	Quantity    int                    `bun:"quantity,notnull"`
	BeforeStock int                    `bun:"before_stock,notnull"`
	AfterStock  int                    `bun:"after_stock,notnull"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
