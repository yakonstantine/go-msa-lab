package repo

import (
	"context"
	"fmt"
	"sync"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type SMTPMemoRepo struct {
	mu sync.RWMutex

	storage map[string]entity.SMTPAddress
}

func NewSMTPMemoRepo() *SMTPMemoRepo {
	return &SMTPMemoRepo{
		storage: map[string]entity.SMTPAddress{},
	}
}

func (sr *SMTPMemoRepo) Create(ctx context.Context, tx usecase.Transaction, smtp *entity.SMTPAddress) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	id := string(smtp.Address)
	if _, ok := sr.storage[id]; ok {
		return fmt.Errorf("smtp address with email '%v' already exists", smtp.Address)
	}

	sr.storage[id] = *smtp
	return nil
}

func (sr *SMTPMemoRepo) GetByEmail(ctx context.Context, eml entity.Email) (*entity.SMTPAddress, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	id := string(eml)
	if smtp, ok := sr.storage[id]; ok {
		return &smtp, nil
	}

	return nil, nil
}

func (sr *SMTPMemoRepo) GetByIdentity(ctx context.Context, identity string) ([]entity.SMTPAddress, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	res := []entity.SMTPAddress{}
	for _, smtp := range sr.storage {
		if smtp.Identity == identity {
			res = append(res, smtp)
		}
	}

	return res, nil
}
