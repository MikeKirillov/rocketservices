package order

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	errs "github.com/MikeKirillov/rocketservices/assembly/internal/errors"
	"github.com/MikeKirillov/rocketservices/assembly/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
	eventsv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/events/v1"
)

const (
	max = 15
	min = 5
)

func (s *service) RunConsumer(ctx context.Context) error {
	slog.InfoContext(ctx, "запуск потребителя OrderPaid")

	return s.orderPaidConsumer.Consume(ctx, s.OrderPaidHandler)
}

func (s *service) OrderPaidHandler(ctx context.Context, msg kafka.Message) error {
	event, err := decodeOrderPaid(msg.Value)
	if err != nil {
		slog.ErrorContext(ctx, "не удалось декодировать OrderPaid", "error", err)
		return err
	}

	slog.InfoContext(
		ctx, "обработка сообщения",
		"topic", msg.Topic,
		"partition", msg.Partition,
		"offset", msg.Offset,
		"order_uuid", event.OrderUuid,
		"user_uuid", event.UserUuid,
	)

	if err := s.shipAssembly(ctx, event); err != nil {
		return err
	}

	return nil
}

func decodeOrderPaid(data []byte) (model.OrderPaidEvent, error) {
	recieved := time.Now()

	var pb eventsv1.OrderPaid
	if err := proto.Unmarshal(data, &pb); err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("не удалось десериализовать protobuf: %w", err)
	}

	orderId, err := uuid.Parse(pb.GetOrderUuid())
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("десериализация %s: %w", pb.GetOrderUuid(), errs.ErrInvalidUUID)
	}

	userId, err := uuid.Parse(pb.GetUserUuid())
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("десериализация %s: %w", pb.GetUserUuid(), errs.ErrInvalidUUID)
	}

	return model.OrderPaidEvent{
		OrderUuid: orderId,
		UserUuid:  userId,
		Recieved:  recieved,
	}, nil
}

func (s *service) shipAssembly(ctx context.Context, event model.OrderPaidEvent) error {
	// Ожидание сборки
	now := time.Now()
	waitTime := time.Duration(rand.IntN(max-min+1)+min) * time.Second //nolint:gosec // таймер с рандомным значением ожидания

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, waitTime)
	defer timeoutCancel()

	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() != nil {
			if timeoutCtx.Err() == context.DeadlineExceeded {
				slog.InfoContext(ctx, "сборка завершена", "время сборки", time.Since(now))
			} else {
				slog.ErrorContext(ctx, "ошибка при сборке", "error", timeoutCtx.Err())
			}
		}
	case <-ctx.Done():
		slog.InfoContext(ctx, "сборка была прервана раньше времени")
	}

	// Отправка события AssemblyShipAssembled в Kafka
	return s.shipAssemblyProducer.ProduceShipAssembly(ctx, event)
}
