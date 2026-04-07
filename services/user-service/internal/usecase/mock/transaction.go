package mock

import (
	"context"

	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/usecase"
)

type Transaction struct{}

func (tx *Transaction) Commit() error {
	return nil
}

func (tx *Transaction) Rollback() {}

type TransactionFactory struct{}

func (tf *TransactionFactory) BeginTx(ctx context.Context) (usecase.Transaction, error) {
	return &Transaction{}, nil
}
