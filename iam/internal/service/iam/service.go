package iam

import "time"

type Service struct {
	userRepository    UserRepository
	sessionRepository SessionRepository

	ttl time.Duration
}

func New(
	userRepository UserRepository,
	sessionRepository SessionRepository,
	ttl time.Duration,
) *Service {
	return &Service{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		ttl:               ttl,
	}
}
