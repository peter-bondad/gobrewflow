package categories

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Category struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID        uuid.UUID `bun:"id,pk" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	IsActive  bool      `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
type CreateCategoryInput struct {
	Name string
}

type CategoryListQuery struct {
	bun.BaseModel `bun:"table:categories,alias:c"`
	Name          string
	IsActive      bool
}

type CategoryList struct {
	bun.BaseModel `bun:"table:categories,alias:c"`
	Name          string
	IsActive      bool
}
