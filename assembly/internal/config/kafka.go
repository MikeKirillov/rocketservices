package config

import "github.com/IBM/sarama"

type kafkaConfig struct {
	Brokers                       []string                            `env:"KAFKA_BROKERS" env-default:"localhost:9092" env-separator:","`
	OrderPaidConsumer             orderPaidConsumerConfig             `yaml:"consumer"`
	AssemblyShipAssembledProducer assemblyShipAssembledProducerConfig `yaml:"producer"`
}

type orderPaidConsumerConfig struct {
	TopicName string `yaml:"topic" env:"ORDER_PAID_TOPIC_NAME" env-default:"order.paid"`
	Group     string `yaml:"group" env:"ORDER_PAID_CONSUMER_GROUP_ID" env-default:"1"`
}

func (c *orderPaidConsumerConfig) Topic() string   { return c.TopicName }
func (c *orderPaidConsumerConfig) GroupID() string { return c.Group }

func (c *orderPaidConsumerConfig) SaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	return cfg
}

type assemblyShipAssembledProducerConfig struct {
	TopicName string `yaml:"topic" env:"ASSEMBLY_SHIP_ASSEMBLED_TOPIC_NAME" env-default:"assembly.ship-assembled"`
}

func (c *assemblyShipAssembledProducerConfig) Topic() string { return c.TopicName }

func (c *assemblyShipAssembledProducerConfig) SaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true
	return cfg
}
