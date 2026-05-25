package order

import (
	"context"

	"github.com/MikeKirillov/rocketservices/assembly/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

type ShipAssemblyProducer interface {
	ProduceShipAssembly(ctx context.Context, order model.OrderPaidEvent) error
}
