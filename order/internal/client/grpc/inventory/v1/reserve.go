package v1

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// ReserveParts резервирует детали для заказа.
func (c *client) ReserveParts(ctx context.Context, partsUuids []uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	_, err := c.grpcClient.ReserveParts(ctx, &inventoryv1.ReservePartsRequest{Uuids: converter.PartUuidsToStr(partsUuids)})
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
