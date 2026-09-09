package categories

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CategoryRepositoryInterface interface {
	InsertCategory(ctx context.Context, category *Category) error
	FindCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	UpdateCategory(ctx context.Context, category *Category) error
	ListCategories(ctx context.Context, params CategoryListParams) (CategoryListResult, error)
}

type categoryRepository struct {
	db *bun.DB
}

func NewCategoryRepository(db *bun.DB) CategoryRepositoryInterface {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) InsertCategory(
	ctx context.Context,
	category *Category,
) error {
	_, err := r.db.NewInsert().
		Model(category).
		Exec(ctx)

	return err
}

func (r *categoryRepository) FindCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	category := new(Category)
	err := r.db.NewSelect().Model(category).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*Category)(nil)).
		Where("name = ?", name).
		Exists(ctx)

	return exists, err
}
func (r *categoryRepository) UpdateCategory(ctx context.Context, category *Category) error {
	_, err := r.db.NewUpdate().Model(category).Where("id = ?", category.ID).Exec(ctx)
	return err
}

type CategoryListParams struct {
	Name   string
	Limit  int
	Offset int
}

type CategoryListResult struct {
	Data  []CategoryList
	Total int
}

func (r *categoryRepository) ListCategories(
	ctx context.Context,
	params CategoryListParams,
) (CategoryListResult, error) {

	categories := make([]CategoryList, 0)

	query := r.db.NewSelect().
		Model(&categories).
		Column("name")

	if params.Name != "" {
		query = query.Where(
			"name ILIKE ?",
			"%"+params.Name+"%",
		)
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return CategoryListResult{}, err
	}

	if err := query.
		Order("name ASC").
		Limit(params.Limit).
		Offset(params.Offset).
		Scan(ctx); err != nil {
		return CategoryListResult{}, err
	}

	return CategoryListResult{
		Data:  categories,
		Total: total,
	}, nil
}
