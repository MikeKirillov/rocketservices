package converter

import (
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func RegisterToInput(proto *userv1.RegisterRequest) input.RegisterInput {
	return input.RegisterInput{
		Login:    proto.GetInfo().GetInfo().GetLogin(),
		Password: proto.GetInfo().GetPassword(),
	}
}
