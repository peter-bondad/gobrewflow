package products

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/uptrace/bun"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, db bun.IDB, product *Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*ProductListItem, error)
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ListProducts(ctx context.Context, params ProductListParams) (*ProductListResult, error)
	UpdateProduct(ctx context.Context, params UpdateProductParams) (*Product, error)
}

type productRepository struct {
	db *bun.DB
}

func NewProductRepository(db *bun.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) InsertProduct(ctx context.Context, db bun.IDB, product *Product) error {
	product.ID = uuid.New()
	_, err := r.db.NewInsert().Model(product).ExcludeColumn("sku").Exec(ctx)
	return err
}

func (r *productRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*ProductListItem, error) {
	product := new(ProductListItem)

	err := r.db.NewSelect().
		Model(product).
		ColumnExpr("p.id").
		ColumnExpr("p.name").
		ColumnExpr("p.sku").
		ColumnExpr("p.description").
		ColumnExpr("p.price").
		ColumnExpr("p.category_id").
		ColumnExpr("p.image_url").
		ColumnExpr("i.quantity").
		Join("JOIN inventory AS i ON i.product_id = p.id").
		Where("p.id = ?", id).
		Where("p.is_active = ?", true).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) FindBySKU(ctx context.Context, sku string) (*Product, error) {
	product := new(Product)
	err := r.db.NewSelect().Model(product).Where("sku = ?", sku).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return product, nil
}
func (r *productRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*Product)(nil)).
		Where("name = ?", name).
		Exists(ctx)

	return exists, err
}

type ProductListParams struct {
	Search   string
	Name     string
	SKU      string
	Category string
	IsActive bool

	MinPrice int64
	MaxPrice int64

	MinQuantity *int64
	MaxQuantity *int64

	Page   int
	Limit  int
	Offset int
}

type ProductListResult struct {
	Data  []ProductListItem
	Total int
}

func (r *productRepository) ListProducts(
	ctx context.Context,
	params ProductListParams,
) (*ProductListResult, error) {
	products := make([]ProductListItem, 0)

	query := r.db.NewSelect().
		Model(&products).
		ColumnExpr("p.id").
		ColumnExpr("p.name").
		ColumnExpr("p.sku").
		ColumnExpr("p.description").
		ColumnExpr("p.price").
		ColumnExpr("p.category_id").
		ColumnExpr("p.image_url").
		ColumnExpr("i.quantity").
		Join("JOIN inventory AS i ON i.product_id = p.id").
		Where("p.is_active = ?", true)

	// General search
	if params.Search != "" {
		search := "%" + params.Search + "%"

		query = query.Where(
			`(
				p.name ILIKE ?
				OR p.sku ILIKE ?
				OR p.description ILIKE ?
			)`,
			search,
			search,
			search,
		)
	}

	// Field-specific filters
	if params.Name != "" {
		query = query.Where(
			"p.name ILIKE ?",
			"%"+params.Name+"%",
		)
	}

	if params.SKU != "" {
		query = query.Where(
			"p.sku ILIKE ?",
			"%"+params.SKU+"%",
		)
	}

	if params.Category != "" {
		query = query.Where(
			"p.category_id = ?",
			params.Category,
		)
	}

	// Price filters
	if params.MinPrice > 0 {
		query = query.Where(
			"p.price >= ?",
			params.MinPrice,
		)
	}

	if params.MaxPrice > 0 {
		query = query.Where(
			"p.price <= ?",
			params.MaxPrice,
		)
	}

	// Inventory quantity filters
	if params.MinQuantity != nil {
		query = query.Where(
			"i.quantity >= ?",
			*params.MinQuantity,
		)
	}

	if params.MaxQuantity != nil {
		query = query.Where(
			"i.quantity <= ?",
			*params.MaxQuantity,
		)
	}

	// Count before pagination
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	// Pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Fetch results
	err = query.
		Order("p.name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &ProductListResult{
		Data:  products,
		Total: total,
	}, nil
}

type UpdateProductParams struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Price       *int64
	IsActive    *bool
	CategoryID  *uuid.UUID
}

func (r *productRepository) UpdateProduct(ctx context.Context, params UpdateProductParams) (*Product, error) {
	product := new(Product)

	query := r.db.NewUpdate().
		Model(product).
		Where("id = ?", params.ID).
		Set("updated_at = NOW()")

	if params.Name != nil {
		query.Set("name = ?", *params.Name)
	}

	if params.Description != nil {
		query.Set("description = ?", *params.Description)
	}

	if params.Price != nil {
		query.Set("price = ?", *params.Price)
	}

	if params.IsActive != nil {
		query.Set("is_active = ?", *params.IsActive)
	}

	if params.CategoryID != nil {
		query.Set("category_id = ?", *params.CategoryID)
	}

	err := query.
		Returning("*").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	return product, nil
}
