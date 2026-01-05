package kafka

import (
	"context"
	"strings"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

type franzClient struct {
	client *kgo.Client
	admin  *kadm.Client
	config models.BrokerConfig
}

type franzClientFactory struct{}

func NewFranzClientFactory() ClientFactory {
	return &franzClientFactory{}
}

func (f *franzClientFactory) NewClient(config models.BrokerConfig) (KafkaClient, error) {
	seeds := strings.Split(config.BootstrapServers, ",")
	for i := range seeds {
		seeds[i] = strings.TrimSpace(seeds[i])
	}

	opts := []kgo.Opt{kgo.SeedBrokers(seeds...)}

	if config.AuthType == models.AuthSASL && config.Username != "" {
		mechanism := plain.Auth{
			User: config.Username,
			Pass: config.Password,
		}.AsMechanism()
		opts = append(opts, kgo.SASL(mechanism))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &franzClient{
		client: client,
		admin:  kadm.NewClient(client),
		config: config,
	}, nil
}

func (c *franzClient) Connect(ctx context.Context) error {
	return c.client.Ping(ctx)
}

func (c *franzClient) Close() {
	c.client.Close()
}

func (c *franzClient) ListTopics(ctx context.Context) ([]models.Topic, error) {
	topics, err := c.admin.ListTopics(ctx)
	if err != nil {
		return nil, err
	}

	topicNames := make([]string, 0, len(topics))
	for name := range topics {
		if !strings.HasPrefix(name, "__") {
			topicNames = append(topicNames, name)
		}
	}

	resourceConfigs, _ := c.admin.DescribeTopicConfigs(ctx, topicNames...)
	configMap := make(map[string]map[string]string)
	for _, rc := range resourceConfigs {
		configMap[rc.Name] = make(map[string]string)
		for _, cfg := range rc.Configs {
			if cfg.Value != nil {
				configMap[rc.Name][cfg.Key] = *cfg.Value
			}
		}
	}

	startOffsets, _ := c.admin.ListStartOffsets(ctx, topicNames...)
	endOffsets, _ := c.admin.ListEndOffsets(ctx, topicNames...)

	result := make([]models.Topic, 0, len(topicNames))
	for _, t := range topics {
		if strings.HasPrefix(t.Topic, "__") {
			continue
		}

		replicas := 0
		totalISR := 0
		urp := 0

		for _, p := range t.Partitions {
			if replicas == 0 {
				replicas = len(p.Replicas)
			}
			totalISR += len(p.ISR)
			if len(p.ISR) < len(p.Replicas) {
				urp++
			}
		}

		var messageCount int64
		for _, p := range t.Partitions {
			startOff := int64(0)
			endOff := int64(0)
			if so, ok := startOffsets.Lookup(t.Topic, p.Partition); ok {
				startOff = so.Offset
			}
			if eo, ok := endOffsets.Lookup(t.Topic, p.Partition); ok {
				endOff = eo.Offset
			}
			messageCount += endOff - startOff
		}

		cleanupPolicy := "delete"
		if cfg, ok := configMap[t.Topic]; ok {
			if val, ok := cfg["cleanup.policy"]; ok {
				cleanupPolicy = val
			}
		}

		result = append(result, models.Topic{
			Name:           t.Topic,
			Partitions:     len(t.Partitions),
			Replicas:       replicas,
			InSyncReplicas: totalISR,
			URP:            urp,
			CleanUpPolicy:  strings.ToUpper(cleanupPolicy),
			MessageCount:   messageCount,
			IsInternal:     t.IsInternal,
		})
	}

	return result, nil
}

func (c *franzClient) GetTopicPartitions(ctx context.Context, topicName string) ([]models.Partition, error) {
	topics, err := c.admin.ListTopics(ctx, topicName)
	if err != nil {
		return nil, err
	}

	topic, ok := topics[topicName]
	if !ok {
		return nil, nil
	}

	startOffsets, err := c.admin.ListStartOffsets(ctx, topicName)
	if err != nil {
		return nil, err
	}

	endOffsets, err := c.admin.ListEndOffsets(ctx, topicName)
	if err != nil {
		return nil, err
	}

	partitions := make([]models.Partition, 0, len(topic.Partitions))
	for _, p := range topic.Partitions {
		startOffset := int64(0)
		endOffset := int64(0)

		if so, ok := startOffsets.Lookup(topicName, p.Partition); ok {
			startOffset = so.Offset
		}
		if eo, ok := endOffsets.Lookup(topicName, p.Partition); ok {
			endOffset = eo.Offset
		}

		partitions = append(partitions, models.Partition{
			ID:             int(p.Partition),
			MessageCount:   endOffset - startOffset,
			StartOffset:    startOffset,
			EndOffset:      endOffset,
			Leader:         int(p.Leader),
			Replicas:       int32SliceToIntSlice(p.Replicas),
			InSyncReplicas: int32SliceToIntSlice(p.ISR),
		})
	}

	return partitions, nil
}

func int32SliceToIntSlice(s []int32) []int {
	result := make([]int, len(s))
	for i, v := range s {
		result[i] = int(v)
	}
	return result
}
