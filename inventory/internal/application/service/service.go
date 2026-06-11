package part

import "context"

// TxManager определяет контракт для управления транзакциями.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type service struct {
	partRepository       PartRepository
	compatibilityChecker CompatibilityChecker
	txManager            TxManager
	authClient           AuthServiceClient
}

func NewService(
	partrepository PartRepository,
	compatibilityChecker CompatibilityChecker,
	txManager TxManager,
	authClient AuthServiceClient,
) *service {
	return &service{
		partRepository:       partrepository,
		compatibilityChecker: compatibilityChecker,
		txManager:            txManager,
		authClient:           authClient,
	}
}
