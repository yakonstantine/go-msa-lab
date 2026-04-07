package repo

import (
	"cmp"
	"context"
	"fmt"
	"slices"
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
		return fmt.Errorf("user with corp key '%v' %w", u.CorpKey, entity.ErrAlreadyExists)
	}

	ur.storage[id] = *u
	return nil
}

func (ur *UserMemoRepo) Update(ctx context.Context, tx usecase.Transaction, u *entity.User) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	id := string(u.CorpKey)
	if _, ok := ur.storage[id]; !ok {
		return fmt.Errorf("user with corp key '%v' %w", u.CorpKey, entity.ErrNotFound)
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

func (ur *UserMemoRepo) GetPage(ctx context.Context, limit, offset int) (entity.Page[entity.User], error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	totalCount := len(ur.storage)
	if offset >= totalCount {
		return entity.Page[entity.User]{
			TotalCount: totalCount,
		}, nil
	}

	end := min(offset+limit, totalCount)

	items := make([]entity.User, 0, totalCount)
	for _, u := range ur.storage {
		items = append(items, u)
	}

	slices.SortFunc(items, func(a, b entity.User) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(string(a.CorpKey), string(b.CorpKey))
	})

	return entity.Page[entity.User]{
		Items:      items[offset:end],
		TotalCount: totalCount,
	}, nil
}
