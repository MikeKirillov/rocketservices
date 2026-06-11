package interceptor

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	invService "github.com/MikeKirillov/rocketservices/inventory/internal/application/service"
	"github.com/MikeKirillov/rocketservices/platform/pkg/auth"
)

// ServerIncomingInterceptor - перехватывает вызов сервера. Извлекает идентификатор сессии
// для дальнейшей валидации пользователя, инициировавшего запрос.
func ServerIncomingInterceptor(authClient invService.AuthServiceClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// Извлекаем метаданные из входящего контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "отсутствуют metadata")
		}

		// Извлекаем идентификатор сессии
		keys := md.Get(auth.SessionUUIDKey)
		if len(keys) == 0 {
			return nil, status.Error(codes.Unauthenticated, "отсутствует session-uuid")
		}

		sessionUUID := keys[0]
		if sessionUUID == "" {
			return nil, status.Error(codes.Unauthenticated, "недействительная сессия")
		}

		sID, uID, err := authClient.Whoami(ctx, uuid.MustParse(sessionUUID))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "ошибка авторизации")
		}

		slog.Info(
			"gRPC: пользователь аутентифицирован",
			"method", info.FullMethod,
			"sessionUUID", sID,
			"userUUID", uID,
		)

		// Передаем обновленный идентификатором пользователя контекст в handler
		return handler(auth.WithUserUUID(ctx, uID), req)
	}
}
