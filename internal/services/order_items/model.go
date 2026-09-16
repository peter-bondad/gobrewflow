package order_items

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OrderItem struct {
	bun.BaseModel `bun:"table:order_items,alias:oi"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrderID   uuid.UUID `bun:"order_id,notnull,type:uuid"`
	ProductID uuid.UUID `bun:"product_id,notnull,type:uuid"`
	Quantity  int       `bun:"quantity,notnull"`
	UnitPrice int64     `bun:"unit_price,notnull"`
	Total     int64     `bun:"total,notnull"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
