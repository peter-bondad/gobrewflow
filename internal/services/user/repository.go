package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// This interface defines the methods that any user repository implementation must provide. It abstracts the underlying data storage mechanism, allowing for different implementations (e.g., in-memory, database) to be swapped out without changing the code that uses the repository.
type UserRepository interface {
	InsertUser(ctx context.Context, tx bun.IDB, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, input UserListInput) ([]UserListItem, error)
	UpdateUser(ctx context.Context, user *User) error
}

// userRepository is a concrete implementation of the UserRepository interface. It uses a *bun.DB instance to interact with the database.
type userRepository struct {
	db bun.IDB
}

// This is like a constructor function for the userRepository struct. It takes a *bun.DB instance as an argument and returns a UserRepository interface. This allows for dependency injection and makes it easier to mock the repository in tests.
func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) InsertUser(ctx context.Context, q bun.IDB, user *User) error {
	_, err := q.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user := new(User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := new(User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) ListUsers(
	ctx context.Context,
	input UserListInput,
) ([]UserListItem, error) {

	userItems := make([]UserListItem, 0)

	query := r.db.NewSelect().
		Model(&userItems).
		Column("id", "email", "full_name", "role").
		Limit(input.Limit).
		Offset(input.Offset)

	if input.UserRole != nil {
		query = query.Where("role = ?", *input.UserRole)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return userItems, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *User) error {
	_, err := r.db.NewUpdate().Model(user).Where("id = ?", user.ID).Exec(ctx)
	return err
}
