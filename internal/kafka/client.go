package kafka

import (
	"context"

	"github.com/jurabek/lazykafka/internal/models"
)

type KafkaClient interface {
	Connect(ctx context.Context) error
	Close()
	ListTopics(ctx context.Context) ([]models.Topic, error)
	GetTopicPartitions(ctx context.Context, topicName string) ([]models.Partition, error)
	CreateTopic(ctx context.Context, config models.TopicConfig) error
	ProduceMessage(ctx context.Context, topic string, key string, value string, headers []models.Header) error
	ConsumeMessages(ctx context.Context, topic string, filter models.MessageFilter) ([]models.Message, error)
	DeleteTopic(ctx context.Context, topicName string) error
	GetTopicConfig(ctx context.Context, topicName string) (models.TopicConfig, error)
	UpdateTopicConfig(ctx context.Context, config models.TopicConfig) error
	GetConsumerGroupOffsets(ctx context.Context, groupName string) ([]models.ConsumerGroupOffset, error)
}

type ClientFactory interface {
	NewClient(config models.BrokerConfig) (KafkaClient, error)
}
