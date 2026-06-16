package iam

import "time"

type service struct {
	userRepository    UserRepository
	sessionRepository SessionRepository

	ttl time.Duration
}

func New(
	userRepository UserRepository,
	sessionRepository SessionRepository,
	ttl time.Duration,
) *service {
	return &service{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		ttl:               ttl,
	}
}
