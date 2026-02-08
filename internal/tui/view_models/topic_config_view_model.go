package viewmodel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jurabek/lazykafka/internal/kafka"
	"github.com/jurabek/lazykafka/internal/models"
)

type TopicConfigViewModel struct {
	mu           sync.RWMutex
	topicName    string
	config       models.TopicConfig
	fieldErrors  map[string]string
	currentField int
	kafkaClient  kafka.KafkaClient
	onError      func(err error)
}

const (
	FieldConfigPartitions = iota
	FieldConfigReplication
	FieldConfigCleanup
	FieldConfigMinSync
	FieldConfigRetention
	FieldConfigCount
)

func NewTopicConfigViewModel(topicName string, config models.TopicConfig, onSave func(models.TopicConfig), onClose func()) *TopicConfigViewModel {
	vm := &TopicConfigViewModel{
		topicName:   topicName,
		config:      config,
		fieldErrors: make(map[string]string),
	}

	return vm
}

func (vm *TopicConfigViewModel) GetName() string {
	return "topic_config"
}

func (vm *TopicConfigViewModel) GetTitle() string {
	return "Topic Config: " + vm.topicName
}

func (vm *TopicConfigViewModel) GetTopicName() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.topicName
}

func (vm *TopicConfigViewModel) GetConfig() models.TopicConfig {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.config
}

func (vm *TopicConfigViewModel) SetConfig(config models.TopicConfig) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.config = config
}

func (vm *TopicConfigViewModel) GetCurrentField() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentField
}

func (vm *TopicConfigViewModel) SetCurrentField(field int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if field >= 0 && field < FieldConfigCount {
		vm.currentField = field
	}
}

func (vm *TopicConfigViewModel) NextField() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.currentField = (vm.currentField + 1) % FieldConfigCount
}

func (vm *TopicConfigViewModel) PrevField() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.currentField--
	if vm.currentField < 0 {
		vm.currentField = FieldConfigCount - 1
	}
}

func (vm *TopicConfigViewModel) GetFieldError(field string) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.fieldErrors[field]
}

func (vm *TopicConfigViewModel) ClearFieldError(field string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	delete(vm.fieldErrors, field)
}

func (vm *TopicConfigViewModel) SetFieldError(field, message string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.fieldErrors[field] = message
}

func (vm *TopicConfigViewModel) SetKafkaClient(client kafka.KafkaClient) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kafkaClient = client
}

func (vm *TopicConfigViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
}

func (vm *TopicConfigViewModel) Validate() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.fieldErrors = make(map[string]string)

	if vm.config.Partitions < 1 {
		vm.fieldErrors["partitions"] = "must be at least 1"
	}

	if vm.config.ReplicationFactor < 1 {
		vm.fieldErrors["replication"] = "must be at least 1"
	}

	if vm.config.MinInSyncReplicas < 1 {
		vm.fieldErrors["minsync"] = "must be at least 1"
	}

	if vm.config.MinInSyncReplicas > vm.config.ReplicationFactor {
		vm.fieldErrors["minsync"] = "cannot exceed replication factor"
	}

	if vm.config.RetentionMs < 0 {
		vm.fieldErrors["retention"] = "must be non-negative"
	}

	if len(vm.fieldErrors) > 0 {
		return ErrValidation
	}

	return nil
}

func (vm *TopicConfigViewModel) LoadFromKafka(topicName string) error {
	vm.mu.RLock()
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("kafka client not configured")
	}

	config, err := client.GetTopicConfig(context.Background(), topicName)
	if err != nil {
		slog.Error("failed to load topic config", slog.String("topic", topicName), slog.Any("error", err))
		if onError != nil {
			onError(err)
		}
		return err
	}

	vm.SetConfig(config)
	return nil
}
