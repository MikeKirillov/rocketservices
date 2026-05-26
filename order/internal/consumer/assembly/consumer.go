package assembly

import (
	"context"
	"errors"
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
	// (лучше эту логику выносить в сервис и в consumer притягивать только его, как было сделано изначально)
	if err := s.commit(ctx, event.OrderUuid); err != nil {
		return err
	}

	slog.Info("заказ обновлен", "order_uuid", event.OrderUuid)

	return nil
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

func (s *service) commit(ctx context.Context, id uuid.UUID) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.GetForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				return sharedErrs.NewNotFound(err)
			}
			return sharedErrs.NewInternal(err)
		}

		// Получить заказ из БД и проверить статус
		// Если заказ уже ASSEMBLED — пропустить (идемпотентность, at-least-once семантика)
		err = order.CanCommit()
		if err != nil {
			return sharedErrs.NewConflict(err)
		}

		// Вызвать CommitParts через gRPC-клиент InventoryService
		err = s.inventoryClient.CommitParts(ctx, order)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrGrpcNotFound):
				return sharedErrs.NewNotFound(err)
			case errors.Is(err, errs.ErrGrpcInvalidArgument):
				return sharedErrs.NewInvalidArgument(err)
			case errors.Is(err, errs.ErrGrpcCancelledOperation):
				return sharedErrs.NewFailedPrecondition(err)
			default:
				return sharedErrs.NewInternal(err)
			}
		}

		// Обновить статус заказа на ASSEMBLED
		order.Commit()
		if err = s.orderRepository.Update(ctx, order); err != nil {
			if errors.Is(err, errs.ErrOrderConflict) {
				return sharedErrs.NewConflict(err)
			}
			return sharedErrs.NewInternal(err)
		}

		return nil
	})
	return err
}
