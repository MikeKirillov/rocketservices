package order

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
	kafkaMdw "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/middleware"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	eventsv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/events/v1"
)

type service struct {
	orderPaidProducer KafkaProducer
}

func NewService(orderPaidProducer KafkaProducer) *service {
	return &service{
		orderPaidProducer: orderPaidProducer,
	}
}

func (s *service) ProduceOrderPaid(ctx context.Context, order model.Order) error {
	msg := &eventsv1.OrderPaid{
		EventUuid: uuid.NewString(),
		OrderUuid: order.OrderUUID.String(),
		UserUuid:  order.UserUUID.String(),
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		slog.ErrorContext(ctx, "не удалось сериализовать OrderPaid", "error", err)
		return sharedErrs.NewInternal(err)
	}

	return s.orderPaidProducer.Send(ctx, &kafka.Message{
		Key:     []byte(order.OrderUUID.String()),
		Value:   payload,
		Headers: kafkaMdw.ProducerSessionHeaders(ctx),
	})
}
