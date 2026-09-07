package categories

import (
	"context"

	"github.com/google/uuid"
)

type CategoryServiceInterface interface {
	CreateCategory(ctx context.Context, name string) (*Category, error)
	UpdateCategoryName(ctx context.Context, id uuid.UUID, name string) (bool, error)
	ListCategories(ctx context.Context, input CategoryListInput) (CategoryListOutput, error)
	SetCategoryStatus(ctx context.Context, id uuid.UUID, isActive bool) (bool, error)
}

type categoryService struct {
	repo CategoryRepositoryInterface
}

func NewCategoryService(repo CategoryRepositoryInterface) CategoryServiceInterface {
	return &categoryService{
		repo: repo,
	}
}

func (s *categoryService) CreateCategory(
	ctx context.Context,
	name string,
) (*Category, error) {
	category := &Category{
		ID:   uuid.New(),
		Name: name,
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *categoryService) UpdateCategoryName(ctx context.Context, id uuid.UUID, name string) (bool, error) {
	category := &Category{
		ID:   id,
		Name: name,
	}

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return false, err
	}

	return true, nil
}

func (s *categoryService) SetCategoryStatus(ctx context.Context, id uuid.UUID, isActive bool) (bool, error) {
	category, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return false, err
	}

	category.IsActive = isActive

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return false, err
	}

	return true, nil
}

type CategoryListInput struct {
	Name  string
	Limit int
	Page  int
}

type CategoryListOutput struct {
	Data       []CategoryList
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

func (s *categoryService) ListCategories(
	ctx context.Context,
	input CategoryListInput,
) (CategoryListOutput, error) {

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

	params := CategoryListParams{
		Name:   input.Name,
		Limit:  input.Limit,
		Offset: offset,
	}

	result, err := s.repo.ListCategories(ctx, params)
	if err != nil {
		return CategoryListOutput{}, err
	}

	totalPages := 0
	if result.Total > 0 {
		totalPages = (result.Total + input.Limit - 1) / input.Limit
	}

	return CategoryListOutput{
		Data:       result.Data,
		Page:       input.Page,
		Limit:      input.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}, nil
}
