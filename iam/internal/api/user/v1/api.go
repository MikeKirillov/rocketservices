package v1

import (
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

type api struct {
	userv1.UnimplementedUserServiceServer

	userService UserService
}

func NewAPI(userService UserService) *api {
	return &api{userService: userService}
}
