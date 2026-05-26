package app

import (
	"context"

	"github.com/IBM/sarama"

	orderConsumer "github.com/MikeKirillov/rocketservices/assembly/internal/consumer/order"
	pr "github.com/MikeKirillov/rocketservices/assembly/internal/producer/order"
	wrappedKafkaConsumer "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/consumer"
	wrappedKafkaProducer "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/producer"
)

type Config struct {
	OrderPaidTopic, ShipAssembledTopic string
	MinBuildTimeSec, MaxBuildTimeSec   int
}

// ConsumerService определяет контракт для запуска Kafka-потребителей.
type ConsumerService interface {
	RunConsumer(ctx context.Context) error
}

func New(producer sarama.SyncProducer, cg sarama.ConsumerGroup, config Config) ConsumerService {
	wc := wrappedKafkaConsumer.NewConsumer(cg, []string{config.OrderPaidTopic})

	wp := wrappedKafkaProducer.NewProducer(producer, config.ShipAssembledTopic)
	p := pr.NewService(wp)
	return orderConsumer.NewService(wc, p)
}
