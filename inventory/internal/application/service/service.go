package part

import (
	"context"

	"github.com/google/uuid"
)

// TxManager определяет контракт для управления транзакциями.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type service struct {
	partRepository       PartRepository
	compatibilityChecker CompatibilityChecker
	txManager            TxManager
}

func NewService(
	partrepository PartRepository,
	compatibilityChecker CompatibilityChecker,
	txManager TxManager,
) *service {
	return &service{
		partRepository:       partrepository,
		compatibilityChecker: compatibilityChecker,
		txManager:            txManager,
	}
}

func uuidsToStringSlice(ids []uuid.UUID) []string {
	strings := make([]string, 0, len(ids))
	for _, id := range ids {
		strings = append(strings, id.String())
	}
	return strings
}
