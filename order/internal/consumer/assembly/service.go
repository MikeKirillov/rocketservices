package assembly

type service struct {
	shipAssemblyConsumer Consumer
	orderRepository      OrderRepository
	inventoryClient      InventoryServiceClient
	txManager            TxManager
}

func NewService(
	shipAssemblyConsumer Consumer,
	orderRepository OrderRepository,
	inventoryClient InventoryServiceClient,
	txManager TxManager,
) *service {
	return &service{
		shipAssemblyConsumer: shipAssemblyConsumer,
		orderRepository:      orderRepository,
		inventoryClient:      inventoryClient,
		txManager:            txManager,
	}
}
