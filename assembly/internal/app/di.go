package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/IBM/sarama"

	"github.com/MikeKirillov/rocketservices/assembly/internal/config"
	orderConsumer "github.com/MikeKirillov/rocketservices/assembly/internal/consumer/order"
	producer "github.com/MikeKirillov/rocketservices/assembly/internal/producer/order"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	wrappedKafkaConsumer "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/consumer"
	kafkaMiddleware "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/middleware"
	wrappedKafkaProducer "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/producer"
)

// ConsumerService определяет контракт для запуска Kafka-потребителей.
type ConsumerService interface {
	RunConsumer(ctx context.Context) error
}

type diContainer struct {
	// Инфраструктура
	consumerGroup sarama.ConsumerGroup
	syncProducer  sarama.SyncProducer

	// Kafka-обёртки
	orderPaidConsumer    *wrappedKafkaConsumer.Consumer
	shipAssemblyProducer *wrappedKafkaProducer.Producer

	// Сервисы
	orderConsumerSvc    ConsumerService
	assemblyProducerSvc orderConsumer.ShipAssemblyProducer
}

// ConsumerGroup возвращает Kafka consumer group
// При первом вызове создаёт группу и регистрирует closer.
func (d *diContainer) ConsumerGroup() sarama.ConsumerGroup {
	if d.consumerGroup == nil {
		consumerGroup, err := sarama.NewConsumerGroup(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().Kafka.OrderPaidConsumer.GroupID(),
			config.AppConfig().Kafka.OrderPaidConsumer.SaramaConfig(),
		)
		if err != nil {
			slog.Error("не удалось создать consumer group", "error", err)
			os.Exit(1)
		}

		closer.Add("Kafka consumer group", func(_ context.Context) error {
			return consumerGroup.Close()
		})

		d.consumerGroup = consumerGroup
	}
	return d.consumerGroup
}

// OrderPaidConsumer возвращает обёртку Kafka-потребителя для событий OrderPaid.
func (d *diContainer) OrderPaidConsumer() *wrappedKafkaConsumer.Consumer {
	if d.orderPaidConsumer == nil {
		d.orderPaidConsumer = wrappedKafkaConsumer.NewConsumer(
			d.ConsumerGroup(),
			[]string{
				config.AppConfig().Kafka.OrderPaidConsumer.Topic(),
			},
			wrappedKafkaConsumer.WithMiddlewares(
				kafkaMiddleware.RecoveryMiddleware(),
				kafkaMiddleware.ConsumerLogging(),
			),
		)
	}
	return d.orderPaidConsumer
}

// SyncProducer возвращает синхронный Kafka-продюсер
// При первом вызове создаёт продюсер и регистрирует closer.
func (d *diContainer) SyncProducer() sarama.SyncProducer {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().Kafka.AssemblyShipAssembledProducer.SaramaConfig(),
		)
		if err != nil {
			slog.Error("не удалось создать sync producer", "error", err)
			os.Exit(1)
		}

		closer.Add("Kafka sync producer", func(_ context.Context) error {
			return p.Close()
		})

		d.syncProducer = p
	}
	return d.syncProducer
}

// ShipAssemblyProducer возвращает обёртку Kafka-потребителя для событий AssemblyShipAssembled.
func (d *diContainer) ShipAssemblyProducer() *wrappedKafkaProducer.Producer {
	if d.shipAssemblyProducer == nil {
		d.shipAssemblyProducer = wrappedKafkaProducer.NewProducer(
			d.SyncProducer(),
			config.AppConfig().Kafka.AssemblyShipAssembledProducer.Topic(),
		)
	}
	return d.shipAssemblyProducer
}

// ProducerService возвращает сервис потребления событий из Kafka.
func (d *diContainer) ProducerService() orderConsumer.ShipAssemblyProducer {
	if d.assemblyProducerSvc == nil {
		d.assemblyProducerSvc = producer.NewService(d.ShipAssemblyProducer())
	}
	return d.assemblyProducerSvc
}

// OrderConsumerService возвращает сервис потребления событий из Kafka.
func (d *diContainer) OrderConsumerService() ConsumerService {
	if d.orderConsumerSvc == nil {
		d.orderConsumerSvc = orderConsumer.NewService(
			d.OrderPaidConsumer(),
			d.ProducerService(),
		)
	}
	return d.orderConsumerSvc
}
