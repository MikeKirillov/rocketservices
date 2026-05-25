package assembly

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/MikeKirillov/rocketservices/order/internal/consumer/assembly/input"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	eventsv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/events/v1"
)

func (s *service) RunConsumer(ctx context.Context) error {
	slog.InfoContext(ctx, "запуск потребителя AssemblyShipAssembled")

	return s.shipAssemblyConsumer.Consume(ctx, s.AssemblyShipAssembledHandler)
}

func (s *service) AssemblyShipAssembledHandler(ctx context.Context, msg kafka.Message) error {
	// Десериализовать событие из protobuf
	event, err := decodeOrderPaid(msg.Value)
	if err != nil {
		slog.ErrorContext(ctx, "не удалось декодировать AssemblyShipAssembled", "error", err)
		return sharedErrs.NewInternal(err)
	}

	slog.InfoContext(
		ctx, "обработка сообщения",
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
		"order_uuid", event.OrderUuid,
		"user_uuid", event.UserUuid,
		"build_time_sec", event.BuildTimeSec,
		"assembled_at", event.AssembledAt,
	)

	// Проверить статус записи и вызвать CommitParts через gRPC-клиент InventoryService
	return s.orderService.Commit(ctx, event.OrderUuid)
}

func decodeOrderPaid(data []byte) (input.ShipAssembledEvent, error) {
	var pb eventsv1.ShipAssembled

	if err := proto.Unmarshal(data, &pb); err != nil {
		return input.ShipAssembledEvent{}, fmt.Errorf("не удалось десериализовать protobuf: %w", err)
	}

	orderId, err := uuid.Parse(pb.GetOrderUuid())
	if err != nil {
		return input.ShipAssembledEvent{}, fmt.Errorf("десериализация %s: %w", pb.GetOrderUuid(), errs.ErrInvalidUUID)
	}

	userId, err := uuid.Parse(pb.GetUserUuid())
	if err != nil {
		return input.ShipAssembledEvent{}, fmt.Errorf("десериализация %s: %w", pb.GetOrderUuid(), errs.ErrInvalidUUID)
	}

	return input.ShipAssembledEvent{
		OrderUuid:    orderId,
		UserUuid:     userId,
		BuildTimeSec: pb.GetBuildTimeSec(),
		AssembledAt:  pb.GetAssembledAt().AsTime(),
	}, nil
}
