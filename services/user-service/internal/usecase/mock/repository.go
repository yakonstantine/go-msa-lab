package mock

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type SMTPRepository struct {
	GetByEmailFn      func(context.Context, entity.Email) (*entity.SMTPAddress, error)
	GetByIdentityFn   func(context.Context, string) ([]entity.SMTPAddress, error)
	GetByIdentitiesFn func(context.Context, []string) (map[string][]entity.SMTPAddress, error)
	CreateFn          func(context.Context, usecase.Transaction, *entity.SMTPAddress) error
	CreateManyFn      func(context.Context, usecase.Transaction, []entity.SMTPAddress) error
	DeleteAllFn       func(context.Context, usecase.Transaction, string) error
}

func (m *SMTPRepository) GetByEmail(ctx context.Context, e entity.Email) (*entity.SMTPAddress, error) {
	return m.GetByEmailFn(ctx, e)
}

func (m *SMTPRepository) GetByIdentity(ctx context.Context, identity string) ([]entity.SMTPAddress, error) {
	return m.GetByIdentityFn(ctx, identity)
}

func (m *SMTPRepository) GetByIdentities(ctx context.Context, identities []string) (map[string][]entity.SMTPAddress, error) {
	return m.GetByIdentitiesFn(ctx, identities)
}

func (m *SMTPRepository) Create(ctx context.Context, tx usecase.Transaction, smtp *entity.SMTPAddress) error {
	return m.CreateFn(ctx, tx, smtp)
}

func (m *SMTPRepository) CreateMany(ctx context.Context, tx usecase.Transaction, smtps []entity.SMTPAddress) error {
	return m.CreateManyFn(ctx, tx, smtps)
}

func (m *SMTPRepository) DeleteAll(ctx context.Context, tx usecase.Transaction, identity string) error {
	return m.DeleteAllFn(ctx, tx, identity)
}

type UserRepository struct {
	GetByCorpKeyFn func(context.Context, entity.CorpKey) (*entity.User, error)
	GetPageFn      func(ctx context.Context, limit, offset int) (entity.Page[entity.User], error)
	CreateFn       func(context.Context, usecase.Transaction, *entity.User) error
	UpdateFn       func(context.Context, usecase.Transaction, *entity.User) error
}

func (m *UserRepository) GetByCorpKey(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
	return m.GetByCorpKeyFn(ctx, ck)
}

func (m *UserRepository) GetPage(ctx context.Context, limit, offset int) (entity.Page[entity.User], error) {
	return m.GetPageFn(ctx, limit, offset)
}

func (m *UserRepository) Create(ctx context.Context, tx usecase.Transaction, u *entity.User) error {
	return m.CreateFn(ctx, tx, u)
}

func (m *UserRepository) Update(ctx context.Context, tx usecase.Transaction, u *entity.User) error {
	return m.UpdateFn(ctx, tx, u)
}
