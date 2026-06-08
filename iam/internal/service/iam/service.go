package iam

type service struct {
	userRepository UserRepository
	// sessionRepository SessionRepository
}

func New(
	userRepository UserRepository,
	// sessionRepository SessionRepository,
) *service {
	return &service{
		userRepository: userRepository,
		// sessionRepository: sessionRepository,
	}
}
