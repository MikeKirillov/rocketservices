package config

import "github.com/IBM/sarama"

// Общая конфигурация Kafka.
type kafkaConfig struct {
	Brokers                       []string                            `env:"KAFKA_BROKERS" env-default:"localhost:9092" env-separator:","`
	OrderPaidProducer             orderPaidProducerConfig             `yaml:"producer"`
	AssemblyShipAssembledConsumer assemblyShipAssembledConsumerConfig `yaml:"consumer"`
}

// Конфигурация продьюсера сообщений сервиса order.
type orderPaidProducerConfig struct {
	TopicName string `yaml:"topic" env:"ORDER_PAID_TOPIC_NAME" env-default:"order.paid"`
}

func (c *orderPaidProducerConfig) Topic() string { return c.TopicName }

func (c *orderPaidProducerConfig) SaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true
	return cfg
}

// Конфигурация консьюмера сообщений сервиса assembly.
type assemblyShipAssembledConsumerConfig struct {
	TopicName string `yaml:"topic" env:"ASSEMBLY_SHIP_ASSEMBLED_TOPIC_NAME" env-default:"assembly.ship-assembled"`
	Group     string `yaml:"group" env:"ASSEMBLY_SHIP_ASSEMBLED_CONSUMER_GROUP_ID" env-default:"1"`
}

func (c *assemblyShipAssembledConsumerConfig) Topic() string { return c.TopicName }

func (c *assemblyShipAssembledConsumerConfig) GroupID() string { return c.Group }

func (c *assemblyShipAssembledConsumerConfig) SaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true
	return cfg
}
