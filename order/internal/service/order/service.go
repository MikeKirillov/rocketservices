package order

type service struct {
	orderRepository OrderRepository
	inventoryClient InventoryServiceClient
	paymentClient   PaymentServiceClient
}

func NewService(
	orderRepository OrderRepository,
	inventoryClient InventoryServiceClient,
	paymentClient PaymentServiceClient,
) *service {
	return &service{
		orderRepository: orderRepository,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
	}
}
