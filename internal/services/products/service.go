package products

import (
	"context"
	"database/sql"
	"errors"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/utils"
)

type ProductInput struct {
	Name       string
	SKU        string
	Slug       string
	CategoryID string
}

type ProductOutput struct {
	Name       string
	SKU        *string
	Slug       string
	CategoryID string
}
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, product *ProductInput) error
	FindByID(ctx context.Context, id string) (*ProductOutput, error)
	FindBySKU(ctx context.Context, sku string) (*ProductOutput, error)
	ListProducts(ctx context.Context, params ProductListInput) (ProductListOutput, error)
}

type productService struct {
	productRepo  ProductRepositoryInterface
	categoryRepo categories.CategoryRepositoryInterface
}

func NewProductService(productRepo ProductRepositoryInterface, categoryRepo categories.CategoryRepositoryInterface) ProductServiceInterface {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, input *ProductInput) error {
	if input.Name == "" {
		return ProductNameIsRequired
	}

	if input.CategoryID == "" {
		return ProductCategoryIDIsRequired
	}

	categoryIDUUID, err := utils.ParseUUID(input.CategoryID)
	if err != nil {
		return err
	}
	category, err := s.categoryRepo.FindCategoryByID(ctx, categoryIDUUID)
	if err != nil {
		return err
	}

	exists, err := s.productRepo.ExistsByName(ctx, input.Name)
	if err != nil {
		return err
	}
	if exists {
		return ProductNameAlreadyExists
	}

	slug := input.Slug
	if slug == "" {
		slug = utils.GenerateSlug(input.Name)
	}

	p := &Product{
		Name:       input.Name,
		Slug:       slug,
		CategoryID: category.ID,
	}

	return s.productRepo.InsertProduct(ctx, p)
}

func (s *productService) FindByID(ctx context.Context, id string) (*ProductOutput, error) {
	if id == "" {
		return nil, ProductIdIsRequired
	}

	uuid, err := utils.ParseUUID(id)
	if err != nil {
		return nil, err
	}
	product, err := s.productRepo.FindByID(ctx, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &ProductOutput{
		Name: product.Name,
		SKU:  product.SKU,
	}, nil
}

func (s *productService) FindBySKU(ctx context.Context, sku string) (*ProductOutput, error) {
	if sku == "" {
		return nil, ProductSKUIsRequired
	}
	product, err := s.productRepo.FindBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &ProductOutput{
		Name: product.Name,
		SKU:  product.SKU,
	}, nil
}

type ProductListInput struct {
	Name     string
	Category string
	IsActive *bool
	Page     int
	Limit    int
	Offset   int
}
type ProductListOutput struct {
	Data       []ProductListItem
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

func (s *productService) ListProducts(ctx context.Context, input ProductListInput) (ProductListOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}

	if input.Limit > 100 {
		input.Limit = 100
	}

	if input.Page <= 0 {
		input.Page = 1
	}

	offset := (input.Page - 1) * input.Limit

	params := ProductListParams{
		Name:   input.Name,
		Limit:  input.Limit,
		Offset: offset,
	}

	result, err := s.productRepo.ListProducts(ctx, params)
	if err != nil {
		return ProductListOutput{}, err
	}

	totalPages := 0
	if result.Total > 0 {
		totalPages = (result.Total + input.Limit - 1) / input.Limit
	}

	return ProductListOutput{
		Data:       result.Data,
		Page:       input.Page,
		Limit:      input.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}, nil
}
