package account

import (
	"context"

	"github.com/uptrace/bun"
)

type AccountRepository interface {
	InsertAccount(ctx context.Context, tx bun.IDB, account *Account) error
}

type accountRepository struct {
	db *bun.DB
}

func NewAccountRepository(db *bun.DB) AccountRepository {
	return &accountRepository{
		db: db,
	}
}

func (r *accountRepository) InsertAccount(ctx context.Context, q bun.IDB, account *Account) error {
	_, err := q.NewInsert().Model(account).Exec(ctx)
	return err
}
