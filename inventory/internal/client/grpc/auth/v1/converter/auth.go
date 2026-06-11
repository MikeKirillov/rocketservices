package converter

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func RqToProto(s uuid.UUID) *authv1.WhoamiRequest {
	return &authv1.WhoamiRequest{
		SessionUuid: s.String(),
	}
}

func ProtoToRs(sessionUUID uuid.UUID, rs *authv1.WhoamiResponse) (uuid.UUID, uuid.UUID, error) {
	if rs.GetSession() == nil || rs.GetSession().GetUuid() == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("сессия %s: %w", sessionUUID, errs.ErrGrpcInvalidSession)
	}

	sID, err := uuid.Parse(rs.GetSession().GetUuid())
	if err != nil {
		slog.Error("неверный формат", "uuid", rs.GetSession().GetUuid())
		return uuid.Nil, uuid.Nil, fmt.Errorf("транзакция %s: %w", rs.GetSession().GetUuid(), errs.ErrInvalidUUID)
	}

	uID, err := uuid.Parse(rs.GetUser().GetUuid())
	if err != nil {
		slog.Error("неверный формат", "uuid", rs.GetUser().GetUuid())
		return uuid.Nil, uuid.Nil, fmt.Errorf("транзакция %s: %w", rs.GetUser().GetUuid(), errs.ErrInvalidUUID)
	}

	return sID, uID, nil
}
