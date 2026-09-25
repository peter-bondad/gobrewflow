package products

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Base model
type Product struct {
	bun.BaseModel `bun:"table:products,alias:p"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name        string    `bun:"name,notnull"`
	SKU         string    `bun:"sku,notnull,unique"`
	Slug        string    `bun:"slug,notnull,unique"`
	Description *string   `bun:"description"`
	Price       int64     `bun:"price,notnull,default:0"`
	IsActive    bool      `bun:"is_active,notnull,default:true"`
	CategoryID  uuid.UUID `bun:"category_id,notnull,type:uuid"`
	ImageURL    *string   `bun:"image_url"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// List product model
type ProductListItem struct {
	bun.BaseModel `bun:"table:products,alias:p"`

	ID          uuid.UUID `bun:"id"`
	Name        string    `bun:"name"`
	SKU         string    `bun:"sku"`
	Description *string   `bun:"description"`
	Price       int64     `bun:"price"`
	CategoryID  uuid.UUID `bun:"category_id"`
	ImageURL    *string   `bun:"image_url"`
	Quantity    int       `bun:"quantity"`
}
