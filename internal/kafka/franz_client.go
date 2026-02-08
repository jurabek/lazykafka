package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

func (c *franzClient) CreateTopic(ctx context.Context, config models.TopicConfig) error {
	configs := map[string]*string{
		"cleanup.policy":      strPtr(config.CleanupPolicy.String()),
		"min.insync.replicas": strPtr(strconv.Itoa(config.MinInSyncReplicas)),
	}

	if config.RetentionMs > 0 {
		configs["retention.ms"] = strPtr(strconv.FormatInt(config.RetentionMs, 10))
	}

	slog.Info("creating topic", slog.String("name", config.Name),
		slog.Int("partitions", config.Partitions),
		slog.Int("replicationFactor", config.ReplicationFactor),
		slog.Any("configs", configs),
	)
	resp, err := c.admin.CreateTopic(ctx, int32(config.Partitions), int16(config.ReplicationFactor), configs, config.Name)
	if err != nil {
		return err
	}
	return resp.Err
}

func strPtr(s string) *string {
	return &s
}

func (c *franzClient) ProduceMessage(ctx context.Context, topic string, key string, value string, headers []models.Header) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: []byte(value),
	}

	if len(headers) > 0 {
		kgoHeaders := make([]kgo.RecordHeader, len(headers))
		for i, h := range headers {
			kgoHeaders[i] = kgo.RecordHeader{
				Key:   h.Key,
				Value: []byte(h.Value),
			}
		}
		record.Headers = kgoHeaders
	}

	if err := c.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		slog.Error("failed to produce message", slog.String("topic", topic), slog.Any("error", err))
		return err
	}

	slog.Info("message produced", slog.String("topic", topic))
	return nil
}

func (c *franzClient) ConsumeMessages(ctx context.Context, topic string, filter models.MessageFilter) ([]models.Message, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	opts := []kgo.Opt{
		kgo.ConsumeTopics(topic),
	}

	if filter.Partition >= 0 {
		partitions := map[string]map[int32]kgo.Offset{
			topic: {
				int32(filter.Partition): kgo.NewOffset(),
			},
		}
		opts = append(opts, kgo.ConsumePartitions(partitions))
	}

	if filter.Offset == 0 {
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	} else if filter.Offset == -1 {
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))
	} else {
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().At(filter.Offset)))
	}

	seeds := strings.Split(c.config.BootstrapServers, ",")
	opts = append(opts, kgo.SeedBrokers(seeds...))

	if c.config.AuthType == models.AuthSASL && c.config.Username != "" {
		mechanism := plain.Auth{
			User: c.config.Username,
			Pass: c.config.Password,
		}.AsMechanism()
		opts = append(opts, kgo.SASL(mechanism))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var messages []models.Message
	seenCount := 0

	for seenCount < limit {
		fetches := client.PollRecords(ctx, limit)
		if fetches.IsClientClosed() {
			break
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			slog.Error("consume errors", slog.Any("errors", errs))
		}

		fetches.EachRecord(func(r *kgo.Record) {
			if seenCount >= limit {
				return
			}

			headers := make([]models.Header, len(r.Headers))
			for i, h := range r.Headers {
				headers[i] = models.Header{
					Key:   h.Key,
					Value: string(h.Value),
				}
			}

			messages = append(messages, models.Message{
				Topic:     r.Topic,
				Partition: int(r.Partition),
				Offset:    r.Offset,
				Key:       string(r.Key),
				Value:     string(r.Value),
				Timestamp: r.Timestamp,
				Headers:   headers,
			})
			seenCount++
		})

		if len(fetches.Records()) == 0 {
			break
		}
	}

	return messages, nil
}

func (c *franzClient) DeleteTopic(ctx context.Context, topicName string) error {
	resps, err := c.admin.DeleteTopics(ctx, topicName)
	if err != nil {
		return err
	}
	if err := resps.Error(); err != nil {
		return err
	}
	slog.Info("topic deleted", slog.String("name", topicName))
	return nil
}

func (c *franzClient) GetTopicConfig(ctx context.Context, topicName string) (models.TopicConfig, error) {
	topicDetails, err := c.admin.ListTopics(ctx, topicName)
	if err != nil {
		return models.TopicConfig{}, err
	}

	topic, ok := topicDetails[topicName]
	if !ok {
		return models.TopicConfig{}, fmt.Errorf("topic %s not found", topicName)
	}

	configs, err := c.admin.DescribeTopicConfigs(ctx, topicName)
	if err != nil {
		return models.TopicConfig{}, err
	}

	config := models.TopicConfig{
		Name:              topicName,
		Partitions:        len(topic.Partitions),
		ReplicationFactor: 0,
		CleanupPolicy:     models.CleanupDelete,
		MinInSyncReplicas: 0,
		RetentionMs:       0,
	}

	if len(topic.Partitions) > 0 {
		config.ReplicationFactor = len(topic.Partitions[0].Replicas)
	}

	for _, rc := range configs {
		for _, cfg := range rc.Configs {
			if cfg.Value == nil {
				continue
			}
			switch cfg.Key {
			case "cleanup.policy":
				switch *cfg.Value {
				case "compact":
					config.CleanupPolicy = models.CleanupCompact
				case "compact,delete":
					config.CleanupPolicy = models.CleanupCompactDelete
				default:
					config.CleanupPolicy = models.CleanupDelete
				}
			case "min.insync.replicas":
				if val, err := strconv.Atoi(*cfg.Value); err == nil {
					config.MinInSyncReplicas = val
				}
			case "retention.ms":
				if val, err := strconv.ParseInt(*cfg.Value, 10, 64); err == nil {
					config.RetentionMs = val
				}
			}
		}
	}

	return config, nil
}

func (c *franzClient) UpdateTopicConfig(ctx context.Context, config models.TopicConfig) error {
	alterConfigs := []kadm.AlterConfig{
		{Name: "cleanup.policy", Value: strPtr(config.CleanupPolicy.String())},
		{Name: "min.insync.replicas", Value: strPtr(strconv.Itoa(config.MinInSyncReplicas))},
	}
	if config.RetentionMs > 0 {
		alterConfigs = append(alterConfigs, kadm.AlterConfig{
			Name:  "retention.ms",
			Value: strPtr(strconv.FormatInt(config.RetentionMs, 10)),
		})
	}

	resps, err := c.admin.AlterTopicConfigs(ctx, alterConfigs, config.Name)
	if err != nil {
		return err
	}
	for _, r := range resps {
		if r.Err != nil {
			return r.Err
		}
	}

	slog.Info("topic config updated", slog.String("name", config.Name))
	return nil
}
