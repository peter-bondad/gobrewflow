package orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Orders struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID          uuid.UUID   `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrderNumber string      `bun:"order_number,notnull,unique"`
	Status      OrderStatus `bun:"status,notnull,default:'pending'"`

	Subtotal int64 `bun:"subtotal,notnull,default:0"`
	Tax      int64 `bun:"tax,notnull,default:0"`
	Discount int64 `bun:"discount,notnull,default:0"`
	Total    int64 `bun:"total,notnull,default:0"`

	CashierID uuid.UUID `bun:"cashier_id,notnull,type:uuid"`

	CompletedAt *time.Time `bun:"completed_at"`
	CancelledAt *time.Time `bun:"cancelled_at"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
