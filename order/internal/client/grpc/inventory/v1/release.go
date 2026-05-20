package v1

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// ReleaseParts освобождает ранее зарезервированные детали.
func (c *client) ReleaseParts(ctx context.Context, order model.Order) error {
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	_, err := c.grpcClient.ReleaseParts(ctx, &inventoryv1.ReleasePartsRequest{Uuids: converter.OrderToPartUuids(order)})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			slog.Error("неизвестная ошибка", "error", err)
			return errs.ErrGrpcInternal
		}
		if err = converter.GrpcErrToDomain(st, err); err != nil {
			return err
		}
	}

	return nil
}
