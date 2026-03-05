package usecase

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
)

type (
	UserRepository interface {
		GetByCorpKey(context.Context, entity.CorpKey) (*entity.User, error)
		Create(context.Context, Transaction, *entity.User) error
	}
	SMTPRepository interface {
		GetByEmail(context.Context, entity.Email) (*entity.SMTPAddress, error)
		GetByIdentity(ctx context.Context, identity string) ([]entity.SMTPAddress, error)
		Create(context.Context, Transaction, *entity.SMTPAddress) error
	}
	Transaction interface {
		Commit() error
		Rollback()
	}
	TransactionFactory interface {
		BeginTx(context.Context) (Transaction, error)
	}
)
