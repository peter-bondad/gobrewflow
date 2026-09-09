package products

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:products,alias:p"`
	ID            uuid.UUID `bun:",pk"`
	SKU           string    `bun:"sku,notnull,unique"`
	Name          string    `bun:"name,notnull"`
	Description   *string   `bun:"description"`
	Price         int64     `bun:"price,notnull,default:0"` // or use decimal type
	StockQuantity int       `bun:"stock_quantity,notnull,default:0"`
	IsActive      bool      `bun:"is_active,notnull,default:true"`
	CategoryID    uuid.UUID `bun:"category_id,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

type ProductListItem struct {
	bun.BaseModel `bun:"table:products,alias:p"`
	Name          string
	SKU           string
	Price         int64
	StockQuantity int
}
