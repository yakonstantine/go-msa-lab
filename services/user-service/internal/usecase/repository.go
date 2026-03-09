package usecase

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
)

type (
	UserRepository interface {
		GetByCorpKey(context.Context, entity.CorpKey) (*entity.User, error)
		GetPage(ctx context.Context, limit, offset int) (entity.Page[entity.User], error)
		Create(context.Context, Transaction, *entity.User) error
		Update(context.Context, Transaction, *entity.User) error
	}
	SMTPRepository interface {
		GetByEmail(context.Context, entity.Email) (*entity.SMTPAddress, error)
		GetByIdentity(context.Context, string) ([]entity.SMTPAddress, error)
		GetByIdentities(context.Context, []string) (map[string][]entity.SMTPAddress, error)
		Create(context.Context, Transaction, *entity.SMTPAddress) error
		CreateMany(context.Context, Transaction, []entity.SMTPAddress) error
		DeleteAll(context.Context, Transaction, string) error
	}
	Transaction interface {
		Commit() error
		Rollback()
	}
	TransactionFactory interface {
		BeginTx(context.Context) (Transaction, error)
	}
)
