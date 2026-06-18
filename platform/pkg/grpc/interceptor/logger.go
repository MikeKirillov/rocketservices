package interceptor

import (
	"context"
	"log/slog"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor создает серверный унарный интерцептор, который логирует
// информацию о времени выполнения методов gRPC сервера.
func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// Извлекаем имя метода из полного пути
		method := path.Base(info.FullMethod)

		// Засекаем время начала выполнения
		start := time.Now()

		// Вызываем обработчик
		resp, err = handler(ctx, req)

		// Вычисляем длительность выполнения
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", method),
			slog.Duration("duration", duration),
		}

		// Форматируем сообщение в зависимости от результата
		if err != nil {
			st, _ := status.FromError(err)
			attrs = append(
				attrs,
				slog.String("code", st.Code().String()),
				slog.String("error", err.Error()),
			)
			slog.LogAttrs(ctx, slog.LevelError, "gRPC метод завершён с ошибкой", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelInfo, "gRPC метод завершён успешно", attrs...)
		}

		return resp, err
	}
}
