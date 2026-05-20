package v1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// ListParts возвращает список деталей.
func (c *client) ListParts(ctx context.Context, partsUuids []uuid.UUID) ([]model.OrderItem, error) {
	// ПАТТЕРН 1: Контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()

	rs, err := c.grpcClient.ListParts(ctx, &inventoryv1.ListPartsRequest{Uuids: converter.PartUuidsToStr(partsUuids)})
	// ПАТТЕРН 2: Обработка gRPC ошибок
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Неизвестная ошибка
			slog.Error("неизвестная ошибка", "error", err)
			return []model.OrderItem{}, errs.ErrGrpcInternal
		}
		if err = converter.GrpcErrToDomain(st, err); err != nil {
			return []model.OrderItem{}, err
		}
	}

	parts := make([]model.OrderItem, len(rs.GetParts()))
	for i, p := range rs.GetParts() {
		part, err := converter.PartToModel(p)
		if err != nil {
			if errors.Is(err, errs.ErrInvalidUUID) {
				slog.Error("неверный формат", "uuid", p.Uuid)
				return []model.OrderItem{}, fmt.Errorf("деталь %s: %w", p.GetUuid(), err)
			}
			if errors.Is(err, errs.ErrWrongPartType) {
				return []model.OrderItem{}, fmt.Errorf("деталь %s: %w", p.GetPartType(), err)
			}
			return []model.OrderItem{}, errs.ErrUnknown
		}
		parts[i] = part
	}
	return parts, nil
}
