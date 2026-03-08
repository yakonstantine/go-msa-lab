package mock

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type SMTPRepository struct {
	GetByEmailFn    func(context.Context, entity.Email) (*entity.SMTPAddress, error)
	GetByIdentityFn func(ctx context.Context, identity string) ([]entity.SMTPAddress, error)
	CreateFn        func(context.Context, usecase.Transaction, *entity.SMTPAddress) error
}

func (m *SMTPRepository) GetByEmail(ctx context.Context, e entity.Email) (*entity.SMTPAddress, error) {
	return m.GetByEmailFn(ctx, e)
}

func (m *SMTPRepository) GetByIdentity(ctx context.Context, identity string) ([]entity.SMTPAddress, error) {
	return m.GetByIdentityFn(ctx, identity)
}

func (m *SMTPRepository) Create(ctx context.Context, tx usecase.Transaction, s *entity.SMTPAddress) error {
	return m.CreateFn(ctx, tx, s)
}
