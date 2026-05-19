package order

import "context"

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type service struct {
	orderRepository OrderRepository
	inventoryClient InventoryServiceClient
	paymentClient   PaymentServiceClient
	txManager       TxManager
}

func NewService(
	orderRepository OrderRepository,
	inventoryClient InventoryServiceClient,
	paymentClient PaymentServiceClient,
	txManager TxManager,
) *service {
	return &service{
		orderRepository: orderRepository,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		txManager:       txManager,
	}
}
