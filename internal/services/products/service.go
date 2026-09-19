package products

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gobrewflow/internal/database"
	"gobrewflow/internal/services/categories"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/utils"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductService interface {
	CreateProducts(ctx context.Context, inputs []CreateProductInput) error
	FindByID(ctx context.Context, id string) (*ProductOutput, error)
	FindBySKU(ctx context.Context, sku string) (*ProductOutput, error)
	ListProducts(ctx context.Context, params ProductListInput) (ProductListOutput, error)
	UpdateProduct(ctx context.Context, input UpdateProductInput) (*UpdateProductOutput, error)
}

type productService struct {
	productRepo      ProductRepository
	categoryRepo     categories.CategoryRepository
	inventoryService inventory.InventoryService
	txManager        database.TxManager
}

func NewProductService(productRepo ProductRepository, categoryRepo categories.CategoryRepository, inventoryService inventory.InventoryService, txManager database.TxManager) ProductService {
	return &productService{
		productRepo:      productRepo,
		categoryRepo:     categoryRepo,
		inventoryService: inventoryService,
		txManager:        txManager,
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

type CreateProductInput struct {
	Name       string
	Slug       string
	Price      int64
	CategoryID string
}

type ProductOutput struct {
	Name       string
	SKU        string
	Slug       string
	CategoryID string
}

func (s *productService) CreateProducts(
	ctx context.Context,
	inputs []CreateProductInput,
) error {
	if len(inputs) == 0 {
		return ErrNoProducts
	}

	return s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		for i := range inputs {
			if err := s.createProduct(ctx, tx, &inputs[i]); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *productService) createProduct(
	ctx context.Context,
	tx bun.IDB,
	input *CreateProductInput,
) error {
	if err := validateCreateProductInput(input); err != nil {
		return err
	}

	categoryID, err := utils.ParseUUID(input.CategoryID)
	if err != nil {
		return err
	}

	category, err := s.categoryRepo.FindCategoryByID(ctx, categoryID)
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

	product := &Product{
		Name:       input.Name,
		Slug:       slug,
		Price:      input.Price,
		CategoryID: category.ID,
	}

	if err := s.productRepo.InsertProduct(ctx, tx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	if err := s.inventoryService.CreateInitialInventory(
		ctx,
		tx,
		product.ID,
	); err != nil {
		return fmt.Errorf("failed to create initial inventory: %w", err)
	}

	return nil
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
	Search   string
	Name     string
	SKU      string
	Category string

	MinPrice int64
	MaxPrice int64

	MinQuantity *int64
	MaxQuantity *int64

	Page  int
	Limit int
}

type ProductListOutput struct {
	Data       []ProductListItem
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

func (s *productService) ListProducts(
	ctx context.Context,
	input ProductListInput,
) (ProductListOutput, error) {
	offset := (input.Page - 1) * input.Limit

	params := ProductListParams{
		Search:      input.Search,
		Name:        input.Name,
		SKU:         input.SKU,
		Category:    input.Category,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		MinQuantity: input.MinQuantity,
		MaxQuantity: input.MaxQuantity,
		Limit:       input.Limit,
		Offset:      offset,
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
