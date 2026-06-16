package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
	commonv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/common/v1"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func RegisterToInput(proto *userv1.RegisterRequest) input.LoginPassInput {
	return input.LoginPassInput{
		Login:    proto.GetInfo().GetInfo().GetLogin(),
		Password: proto.GetInfo().GetPassword(),
	}
}

func LoginToInput(proto *authv1.LoginRequest) input.LoginPassInput {
	return input.LoginPassInput{
		Login:    proto.GetLogin(),
		Password: proto.GetPassword(),
	}
}

func UserToProto(user model.User) *userv1.GetUserResponse {
	result := &userv1.GetUserResponse{
		User: userToCommonProto(user),
	}

	if user.UpdatedAt != nil {
		result.User.UpdatedAt = timestamppb.New(*user.UpdatedAt)
	}

	return result
}

func UserSessionToProto(user model.User, session model.Session) *authv1.WhoamiResponse {
	return &authv1.WhoamiResponse{
		Session: sessionToCommonProto(session),
		User:    userToCommonProto(user),
	}
}

func userToCommonProto(user model.User) *commonv1.User {
	return &commonv1.User{
		Uuid: user.UUID.String(),
		Info: &commonv1.UserInfo{
			Login: user.Login,
		},
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}

func sessionToCommonProto(session model.Session) *commonv1.Session {
	result := &commonv1.Session{
		Uuid:      session.UUID.String(),
		CreatedAt: timestamppb.New(session.CreatedAt),
		ExpiresAt: timestamppb.New(session.ExpiresAt),
	}

	if session.UpdatedAt != nil {
		result.UpdatedAt = timestamppb.New(*session.UpdatedAt)
	}

	return result
}
