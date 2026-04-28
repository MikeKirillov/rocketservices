package v1

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/payment/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

const (
	// Таймаут для gRPC вызовов.
	grpcCallTimeout = 5 * time.Second
)

type client struct {
	grpcClient paymentv1.PaymentServiceClient
}

// New создаёт обёртку над gRPC клиентом PaymentService.
func New(grpcClient paymentv1.PaymentServiceClient) *client {
	return &client{grpcClient: grpcClient}
}

// PayOrder обрабатывает оплату заказа.
func (c *client) PayOrder(ctx context.Context, orderRq model.PayOrderRq) (string, error) {
	// ПАТТЕРН 1: Контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	rs, err := c.grpcClient.PayOrder(ctx, converter.PaymentToProto(orderRq))
	// ПАТТЕРН 2: Обработка gRPC ошибок
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Неизвестная ошибка
			slog.Error("неизвестная ошибка", "error", err)
			return "", errs.ErrGrpcInternal
		}
		// Маппинг gRPC ошибок в доменные ошибки
		switch st.Code() {
		case codes.NotFound:
			slog.Error("обработать платеж", "message", st.Message())
			return "", errs.ErrInventoryNotFound

		case codes.InvalidArgument:
			slog.Error("обработать платеж", "message", st.Message())
			return "", errs.ErrGrpcInvalidArgument

		case codes.DeadlineExceeded: // Таймаут — клиент может повторить запрос
			slog.Error("таймаут", "message", st.Message())
			return "", errs.ErrGrpcTimeout

		case codes.Unavailable: // gRPC сервис недоступен
			slog.Error("gRPC сервис недоступен", "message", st.Message())
			return "", errs.ErrGrpcUnavailable

		case codes.Canceled: // Запрос был отменён (например, клиент закрыл соединение)
			slog.Error("запрос отменен", "message", st.Message())
			return "", errs.ErrGrpcUnavailable

		default:
			slog.Error("ошибка gRPC", "code", st.Code(), "message", st.Message())
			return "", errs.ErrGrpcInternal
		}
	}

	return rs.GetTransactionUuid(), nil
}
