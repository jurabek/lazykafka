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
}

type ClientFactory interface {
	NewClient(config models.BrokerConfig) (KafkaClient, error)
}
