package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	commonv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/common/v1"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func RegisterToInput(proto *userv1.RegisterRequest) input.RegisterInput {
	return input.RegisterInput{
		Login:    proto.GetInfo().GetInfo().GetLogin(),
		Password: proto.GetInfo().GetPassword(),
	}
}

func UserToProto(user model.User) *userv1.GetUserResponse {
	result := &userv1.GetUserResponse{
		User: &commonv1.User{
			Uuid: user.UUID.String(),
			Info: &commonv1.UserInfo{
				Login: user.Login,
			},
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}

	if user.UpdatetAt != nil {
		result.User.UpdatedAt = timestamppb.New(*user.UpdatetAt)
	}

	return result
}
