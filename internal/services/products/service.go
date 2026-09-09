package products

import (
	"context"
	"gobrewflow/internal/utils"
)

type ProductInput struct {
	Name string
	SKU  string
}

type ProductOutput struct {
	Name string
	SKU  string
}
type ProductServiceInterface interface {
	InsertProduct(ctx context.Context, product *ProductInput) error
	FindByID(ctx context.Context, id string) (*ProductOutput, error)
	FindBySKU(ctx context.Context, sku string) (*ProductOutput, error)
	ListProducts(ctx context.Context, params ProductListParams) (ProductListOutput, error)
}

type productService struct {
	repo ProductRepositoryInterface
}

func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface {
	return &productService{
		repo: repo,
	}
}
func (s *productService) InsertProduct(ctx context.Context, input *ProductInput) error {
	if input == nil {
		return ProductInputIsRequired
	}

	if input.Name == "" {
		return ProductNameIsRequired
	}

	sku := input.SKU
	if sku == "" {
		sku = utils.GenerateSlug(input.Name)
	}

	p := &Product{
		Name: input.Name,
		SKU:  sku,
	}

	return s.repo.InsertProduct(ctx, p)
}

func (s *productService) FindByID(ctx context.Context, id string) (*ProductOutput, error) {
	if id == "" {
		return nil, ProductIdIsRequired
	}
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
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
	product, err := s.repo.FindBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	return &ProductOutput{
		Name: product.Name,
		SKU:  product.SKU,
	}, nil
}

type ProductListOutput struct {
	Data       []ProductListItem
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

func (s *productService) ListProducts(ctx context.Context, input ProductListParams) (ProductListOutput, error) {
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
		SKU:    input.SKU,
		Limit:  input.Limit,
		Offset: offset,
	}

	result, err := s.repo.ListProducts(ctx, params)
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
