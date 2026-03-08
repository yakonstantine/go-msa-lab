package repo

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type TransactionMemo struct{}

func (tx *TransactionMemo) Commit() error {
	return nil
}

func (tx *TransactionMemo) Rollback() {}

type TransactionFactoryMemo struct{}

func (tf *TransactionFactoryMemo) BeginTx(ctx context.Context) (usecase.Transaction, error) {
	return &TransactionMemo{}, nil
}
