package assembly

type service struct {
	shipAssemblyConsumer Consumer
}

func NewService(shipAssemblyConsumer Consumer) *service {
	return &service{
		shipAssemblyConsumer: shipAssemblyConsumer,
	}
}
