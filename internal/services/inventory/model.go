package inventory

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Inventory struct {
	bun.BaseModel `bun:"table:inventory,alias:i"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ProductID uuid.UUID `bun:"product_id,notnull,type:uuid,unique"`
	Quantity  int       `bun:"quantity,notnull,default:0"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
