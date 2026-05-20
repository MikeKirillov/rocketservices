package v1

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

// ValidateCompatibility проверяет совместимость деталей между собой.
func (c *client) ValidateCompatibility(ctx context.Context, items []model.OrderItem) error {
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	_, err := c.grpcClient.ValidateCompatibility(ctx, converter.ItemsToProto(items))
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			slog.Error("неизвестная ошибка", "error", err)
			return errs.ErrGrpcInternal
		}
		if err = converter.GrpcErrToDomain(st); err != nil {
			return err
		}
	}

	return nil
}
