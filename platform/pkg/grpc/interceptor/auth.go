package interceptor

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/platform/pkg/auth"
)

// OutgoingClientInterceptor - осуществляет перехват вызовов с клиентской стороны. Получает
// идентификатор сессии из текущего контекста и передает его в исходящий контекст RPC.
func OutgoingClientInterceptor(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()

	// Извлекаем идентификатор сессии из контекста
	sessionUUID, ok := auth.SessionUUIDFromContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "отсутствует session-uuid")
	}

	// Устанавливаем его в исходящий контекст
	metadata.AppendToOutgoingContext(ctx, auth.SessionUUIDKey, sessionUUID)

	// Осуществляем вызов RPC
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		slog.Error(
			"gRPC метод завершён с ошибкой",
			"method", method,
			"error", err,
			"duration", time.Since(start),
		)
	} else {
		slog.Info(
			"gRPC метод завершён успешно",
			"method", method,
			"duration", time.Since(start),
		)
	}
	return err
}
