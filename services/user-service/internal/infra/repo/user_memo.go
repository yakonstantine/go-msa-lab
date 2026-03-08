package repo

import (
	"context"
	"fmt"
	"sync"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/entity"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type UserMemoRepo struct {
	mu sync.RWMutex

	storage map[string]entity.User
}

func NewUserMemoRepo() *UserMemoRepo {
	return &UserMemoRepo{
		storage: map[string]entity.User{},
	}
}

func (ur *UserMemoRepo) Create(ctx context.Context, tx usecase.Transaction, u *entity.User) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	id := string(u.CorpKey)
	if _, ok := ur.storage[id]; ok {
		return fmt.Errorf("user with corp key '%v' already exists", u.CorpKey)
	}

	ur.storage[id] = *u
	return nil
}

func (ur *UserMemoRepo) GetByCorpKey(ctx context.Context, ck entity.CorpKey) (*entity.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	id := string(ck)
	if u, ok := ur.storage[id]; ok {
		return &u, nil
	}

	return nil, nil
}
