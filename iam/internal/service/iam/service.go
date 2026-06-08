package iam

import (
	"context"
)

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

func (s *service) Login(context.Context) error {
	return nil
}

func (s *service) Whoami(context.Context) error {
	return nil
}

func (s *service) Logout(context.Context) error {
	return nil
}
