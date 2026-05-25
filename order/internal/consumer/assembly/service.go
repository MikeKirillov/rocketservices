package assembly

type service struct {
	shipAssemblyConsumer Consumer
	orderService         OrderService
}

func NewService(
	shipAssemblyConsumer Consumer,
	orderService OrderService,
) *service {
	return &service{
		shipAssemblyConsumer: shipAssemblyConsumer,
		orderService:         orderService,
	}
}
