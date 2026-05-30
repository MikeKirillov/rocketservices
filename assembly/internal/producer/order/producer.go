package order

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MikeKirillov/rocketservices/assembly/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
	eventsv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/events/v1"
)

type service struct {
	shipAssemblyProducer KafkaProducer
}

func NewService(shipAssemblyProducer KafkaProducer) *service {
	return &service{
		shipAssemblyProducer: shipAssemblyProducer,
	}
}

func (s *service) ProduceShipAssembly(ctx context.Context, order model.OrderPaidEvent) error {
	msg := &eventsv1.ShipAssembled{
		EventUuid:    uuid.NewString(),
		OrderUuid:    order.OrderUuid.String(),
		UserUuid:     order.UserUuid.String(),
		BuildTimeSec: int64(time.Since(order.Received).Seconds()),
		AssembledAt:  timestamppb.Now(),
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		slog.ErrorContext(ctx, "не удалось сериализовать AssemblyShipAssembled", "error", err)
		return err
	}

	return s.shipAssemblyProducer.Send(ctx, &kafka.Message{
		Key:   []byte(order.OrderUuid.String()),
		Value: payload,
	})
}
