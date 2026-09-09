package products

import (
	"context"

	"github.com/google/uuid"

	"github.com/uptrace/bun"
)

type ProductRepositoryInterface interface {
	InsertProduct(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id string) (*Product, error)
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	ListProducts(ctx context.Context, params ProductListParams) (*ProductListResult, error)
}

type productRepository struct {
	db bun.DB
}

func NewProductRepository(db bun.DB) ProductRepositoryInterface {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) InsertProduct(ctx context.Context, product *Product) error {
	product.ID = uuid.New()
	_, err := r.db.NewInsert().Model(product).Exec(ctx)
	return err
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*Product, error) {
	product := new(Product)
	err := r.db.NewSelect().Model(product).Where("id = ?", id).Scan(ctx)
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

type ProductListParams struct {
	Name     string
	SKU      string
	Category string
	IsActive *bool
	Limit    int
	Offset   int
}

type ProductListResult struct {
	Data  []ProductListItem
	Total int
}

func (r *productRepository) ListProducts(ctx context.Context, params ProductListParams) (*ProductListResult, error) {
	products := make([]ProductListItem, 0) // Initialize an empty slice to hold the products instead of a pointer to a slice

	query := r.db.NewSelect().Model(&products)

	if params.Name != "" {
		query = query.Where("name ILIKE ?", "%"+params.Name+"%")
	}

	if params.SKU != "" {
		query = query.Where("sku ILIKE ?", "%"+params.SKU+"%")
	}

	if params.Category != "" {
		query = query.Where("category_id = ?", params.Category)
	}

	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return &ProductListResult{}, err
	}

	err = query.Order("name ASC").
		Limit(params.Limit).
		Offset(params.Offset).
		Scan(ctx)
	if err != nil {
		return &ProductListResult{}, err
	}

	return &ProductListResult{
		Data:  products,
		Total: total,
	}, nil

}
