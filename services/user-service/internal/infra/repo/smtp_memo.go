package repo

import (
	"context"
	"fmt"
	"slices"
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
		return fmt.Errorf("smtp address with email '%v' %w", smtp.Address, entity.ErrAlreadyExists)
	}

	sr.storage[id] = *smtp
	return nil
}

func (sr *SMTPMemoRepo) CreateMany(ctx context.Context, tx usecase.Transaction, smtps []entity.SMTPAddress) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	existingSMTPs := []string{}
	for _, smtp := range smtps {
		id := string(smtp.Address)
		if _, ok := sr.storage[id]; ok {
			existingSMTPs = append(existingSMTPs, id)
		}
	}
	if len(existingSMTPs) > 0 {
		return fmt.Errorf("smtp addresses with email '%v' %w", smtps, entity.ErrAlreadyExists)
	}

	for _, smtp := range smtps {
		id := string(smtp.Address)
		sr.storage[id] = smtp
	}
	return nil
}

func (sr *SMTPMemoRepo) DeleteAll(ctx context.Context, tx usecase.Transaction, identity string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	var keysToDelete []string
	for k, smtp := range sr.storage {
		if smtp.Identity == identity {
			keysToDelete = append(keysToDelete, k)
		}
	}

	for _, k2d := range keysToDelete {
		delete(sr.storage, k2d)
	}

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

func (sr *SMTPMemoRepo) GetByIdentities(ctx context.Context, identities []string) (map[string][]entity.SMTPAddress, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	res := map[string][]entity.SMTPAddress{}
	for _, smtp := range sr.storage {
		if slices.Contains(identities, smtp.Identity) {
			res[smtp.Identity] = append(res[smtp.Identity], smtp)
		}
	}

	return res, nil
}
