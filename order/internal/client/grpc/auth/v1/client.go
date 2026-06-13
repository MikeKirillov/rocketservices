package v1

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/auth/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

const grpcCallTimeout = 5 * time.Second

type Client struct {
	grpcClient authv1.AuthServiceClient
}

func New(grpcClient authv1.AuthServiceClient) *Client {
	return &Client{grpcClient: grpcClient}
}

func (c *Client) Whoami(ctx context.Context, sessionUUID string) (uuid.UUID, uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	rs, err := c.grpcClient.Whoami(ctx, &authv1.WhoamiRequest{SessionUuid: sessionUUID})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Неизвестная ошибка
			return uuid.Nil, uuid.Nil, errs.ErrGrpcInternal
		}
		// Маппинг gRPC ошибок в доменные ошибки
		switch st.Code() {
		case codes.NotFound:
			slog.Error("обработать платеж", "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcNotFound

		case codes.InvalidArgument:
			slog.Error("обработать платеж", "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcInvalidArgument

		case codes.DeadlineExceeded: // Таймаут — клиент может повторить запрос
			slog.Error("таймаут", "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcTimeout

		case codes.Unavailable: // gRPC сервис недоступен
			slog.Error("gRPC сервис недоступен", "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcUnavailable

		case codes.Canceled: // Запрос был отменён (например, клиент закрыл соединение)
			slog.Error("запрос отменен", "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcUnavailable

		default:
			slog.Error("ошибка gRPC", "code", st.Code(), "message", st.Message())
			return uuid.Nil, uuid.Nil, errs.ErrGrpcInternal
		}
	}

	sID, uID, err := converter.ProtoToRs(sessionUUID, rs)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return sID, uID, nil
}
