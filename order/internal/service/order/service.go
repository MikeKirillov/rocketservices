package order

import "context"

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type service struct {
	orderRepository      OrderRepository
	inventoryClient      InventoryServiceClient
	paymentClient        PaymentServiceClient
	txManager            TxManager
	orderProducerService OrderProducerService
}

func NewService(
	orderRepository OrderRepository,
	inventoryClient InventoryServiceClient,
	paymentClient PaymentServiceClient,
	txManager TxManager,
	orderProducerService OrderProducerService,
) *service {
	return &service{
		orderRepository:      orderRepository,
		inventoryClient:      inventoryClient,
		paymentClient:        paymentClient,
		txManager:            txManager,
		orderProducerService: orderProducerService,
	}
}
