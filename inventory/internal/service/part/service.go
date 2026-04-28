package part

type service struct {
	partRepository PartRepository
}

func NewService(partrepository PartRepository) *service {
	return &service{partRepository: partrepository}
}
