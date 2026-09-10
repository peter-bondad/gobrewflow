package products

import (
	"context"
	"database/sql"
	"errors"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/utils"
	"strings"

	"github.com/google/uuid"
)

type CreateProductInput struct {
	Name       string
	SKU        string
	Slug       string
	CategoryID string
}

type ProductOutput struct {
	Name       string
	SKU        string
	Slug       string
	CategoryID string
}
type ProductServiceInterface interface {
	CreateProduct(ctx context.Context, product *CreateProductInput) error
	FindByID(ctx context.Context, id string) (*ProductOutput, error)
	FindBySKU(ctx context.Context, sku string) (*ProductOutput, error)
	ListProducts(ctx context.Context, params ProductListInput) (ProductListOutput, error)
	UpdateProduct(ctx context.Context, input UpdateProductInput) (*UpdateProductOutput, error)
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

func validateCreateProductInput(input *CreateProductInput) error {
	if input.Name != "" {
		name := strings.TrimSpace(input.Name)

		if name == "" {
			return ProductNameCannotBeEmpty
		}

		if len(name) > 255 {
			return ProductNameTooLong
		}
	}
	return nil
}
func (s *productService) CreateProduct(ctx context.Context, input *CreateProductInput) error {
	if err := validateCreateProductInput(input); err != nil {
		return err
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
		Name:     input.Name,
		Category: input.Category,
		IsActive: input.IsActive,
		Limit:    input.Limit,
		Offset:   offset,
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

type UpdateProductInput struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Price       *int64
	IsActive    *bool
	CategoryID  *uuid.UUID
}

type UpdateProductOutput struct {
	Product *Product
}

func (s *productService) UpdateProduct(ctx context.Context, input UpdateProductInput) (*UpdateProductOutput, error) {

	if err := validateUpdateProductInput(input); err != nil {
		return nil, err
	}

	product, err := s.productRepo.UpdateProduct(
		ctx,
		UpdateProductParams{
			ID:          input.ID,
			Name:        input.Name,
			Description: input.Description,
			Price:       input.Price,
			IsActive:    input.IsActive,
			CategoryID:  input.CategoryID,
		},
	)
	if err != nil {
		return nil, err
	}

	return &UpdateProductOutput{
		Product: product,
	}, nil
}

func validateUpdateProductInput(input UpdateProductInput) error {
	if input.ID == uuid.Nil {
		return errors.New("product id is required")
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)

		if name == "" {
			return ProductNameCannotBeEmpty
		}

		if len(name) > 255 {
			return ProductNameTooLong
		}
	}

	if input.Price != nil && *input.Price < 0 {
		return ProductPriceNotNegative
	}

	return nil
}
