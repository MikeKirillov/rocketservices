package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

func RecoveryMiddleware() kafka.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "паника при обработке сообщения Kafka", "topic", msg.Topic)
					slog.ErrorContext(
						ctx, "паника при обработке сообщения Kafka",
						"topic", msg.Topic,
						"panic", r,
						"stack", string(debug.Stack()),
					)
				}
			}()
			return next(ctx, msg)
		}
	}
}
