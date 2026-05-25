package order

type service struct {
	orderPaidConsumer    Consumer
	shipAssemblyProducer ShipAssemblyProducer
}

func NewService(
	orderPaidConsumer Consumer,
	shipAssemblyProducer ShipAssemblyProducer,
) *service {
	return &service{
		orderPaidConsumer:    orderPaidConsumer,
		shipAssemblyProducer: shipAssemblyProducer,
	}
}
