package viewmodel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jurabek/lazykafka/internal/kafka"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type MainViewModel struct {
	mu sync.RWMutex

	brokersVM              *BrokersViewModel
	topicsVM               *TopicsViewModel
	consumerGroupsVM       *ConsumerGroupsViewModel
	schemaRegistryVM       *SchemaRegistryViewModel
	topicDetailVM          *TopicDetailViewModel
	messageBrowserVM       *MessageBrowserViewModel
	messageDetailVM        *MessageDetailViewModel
	produceMessageVM       *ProduceMessageViewModel
	consumerGroupDetailVM  *ConsumerGroupDetailViewModel
	schemaRegistryDetailVM *SchemaRegistryDetailViewModel

	onChange types.OnChangeFunc
	ctx      context.Context

	clientFactory kafka.ClientFactory
	activeClient  kafka.KafkaClient
	brokerConfigs []models.BrokerConfig
	onError       func(err error)
}

func NewMainViewModel(
	ctx context.Context,
	configs []models.BrokerConfig,
	factory kafka.ClientFactory,
) *MainViewModel {
	vm := &MainViewModel{
		brokersVM:              NewBrokersViewModel(),
		topicsVM:               NewTopicsViewModel(),
		consumerGroupsVM:       NewConsumerGroupsViewModel(),
		schemaRegistryVM:       NewSchemaRegistryViewModel(),
		topicDetailVM:          NewTopicDetailViewModel(),
		messageBrowserVM:       NewMessageBrowserViewModel(),
		messageDetailVM:        NewMessageDetailViewModel(),
		produceMessageVM:       NewProduceMessageViewModel(nil, nil),
		consumerGroupDetailVM:  NewConsumerGroupDetailViewModel(),
		schemaRegistryDetailVM: NewSchemaRegistryDetailViewModel(),
		ctx:                    ctx,
		clientFactory:          factory,
		brokerConfigs:          configs,
	}

	vm.setupBrokerSelectionCallback()
	vm.setupTopicSelectionCallback()
	vm.setupConsumerGroupSelectionCallback()
	vm.setupSchemaRegistrySelectionCallback()

	return vm
}

func (vm *MainViewModel) LoadInitialData() {
	brokers := configsToBrokers(vm.brokerConfigs)
	vm.brokersVM.Load(brokers)
}

func configsToBrokers(configs []models.BrokerConfig) []models.Broker {
	brokers := make([]models.Broker, len(configs))
	for i, c := range configs {
		brokers[i] = models.Broker{
			ID:      i,
			Name:    c.Name,
			Address: c.BootstrapServers,
		}
	}
	return brokers
}

func (vm *MainViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
	vm.topicsVM.SetOnError(fn)
	vm.topicDetailVM.SetOnError(fn)
	vm.messageBrowserVM.SetOnError(fn)
}

func (vm *MainViewModel) setupBrokerSelectionCallback() {
	vm.brokersVM.SetOnSelectionChanged(func(broker *models.Broker) {
		slog.Info("broker selection changed", slog.String("broker", broker.Name))
		vm.loadDependentData(broker)
	})
}

// setupTopicSelectionCallback registers callback for topic selection changes
func (vm *MainViewModel) setupTopicSelectionCallback() {
	vm.topicsVM.SetOnSelectionChanged(func(topic *models.Topic) {
		vm.topicDetailVM.SetTopic(topic)
	})
}

// setupConsumerGroupSelectionCallback registers callback for consumer group selection changes
func (vm *MainViewModel) setupConsumerGroupSelectionCallback() {
	vm.consumerGroupsVM.SetOnSelectionChanged(func(cg *models.ConsumerGroup) {
		vm.consumerGroupDetailVM.SetConsumerGroup(cg)
	})
}

// setupSchemaRegistrySelectionCallback registers callback for schema registry selection changes
func (vm *MainViewModel) setupSchemaRegistrySelectionCallback() {
	vm.schemaRegistryVM.SetOnSelectionChanged(func(sr *models.SchemaRegistry) {
		vm.schemaRegistryDetailVM.SetSchema(sr)
	})
}

func (vm *MainViewModel) getConfigForBroker(broker *models.Broker) *models.BrokerConfig {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	for i := range vm.brokerConfigs {
		if vm.brokerConfigs[i].Name == broker.Name {
			return &vm.brokerConfigs[i]
		}
	}
	return nil
}

// loadDependentData triggers async reload of all dependent ViewModels
func (vm *MainViewModel) loadDependentData(broker *models.Broker) {
	vm.mu.Lock()
	if vm.activeClient != nil {
		vm.activeClient.Close()
		vm.activeClient = nil
	}
	factory := vm.clientFactory
	onError := vm.onError
	vm.mu.Unlock()

	config := vm.getConfigForBroker(broker)
	if config == nil || factory == nil {
		return
	}

	client, err := factory.NewClient(*config)
	if err != nil {
		slog.Error("failed to create kafka client", slog.Any("error", err))
		if onError != nil {
			onError(err)
		}
		return
	}

	if err := client.Connect(vm.ctx); err != nil {
		slog.Error("failed to connect to kafka", slog.Any("error", err))
		if onError != nil {
			onError(err)
		}
		client.Close()
		return
	}

	vm.mu.Lock()
	vm.activeClient = client
	vm.mu.Unlock()

	vm.topicsVM.SetKafkaClient(client)
	vm.topicDetailVM.SetKafkaClient(client)
	vm.messageBrowserVM.SetKafkaClient(client)

	vm.topicsVM.LoadForBroker(broker)
	vm.consumerGroupsVM.LoadForBroker(broker)
	vm.schemaRegistryVM.LoadForBroker(broker)
}

func (vm *MainViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *MainViewModel) BrokersVM() *BrokersViewModel {
	return vm.brokersVM
}

func (vm *MainViewModel) TopicsVM() *TopicsViewModel {
	return vm.topicsVM
}

func (vm *MainViewModel) ConsumerGroupsVM() *ConsumerGroupsViewModel {
	return vm.consumerGroupsVM
}

func (vm *MainViewModel) SchemaRegistryVM() *SchemaRegistryViewModel {
	return vm.schemaRegistryVM
}

func (vm *MainViewModel) TopicDetailVM() *TopicDetailViewModel {
	return vm.topicDetailVM
}

func (vm *MainViewModel) MessageBrowserVM() *MessageBrowserViewModel {
	return vm.messageBrowserVM
}

func (vm *MainViewModel) MessageDetailVM() *MessageDetailViewModel {
	return vm.messageDetailVM
}

func (vm *MainViewModel) ProduceMessageVM() *ProduceMessageViewModel {
	return vm.produceMessageVM
}

func (vm *MainViewModel) ConsumerGroupDetailVM() *ConsumerGroupDetailViewModel {
	return vm.consumerGroupDetailVM
}

func (vm *MainViewModel) SchemaRegistryDetailVM() *SchemaRegistryDetailViewModel {
	return vm.schemaRegistryDetailVM
}

func (vm *MainViewModel) AddBrokerConfig(config models.BrokerConfig) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.brokerConfigs = append(vm.brokerConfigs, config)
}

func (vm *MainViewModel) CreateTopic(ctx context.Context, config models.TopicConfig) error {
	vm.mu.RLock()
	client := vm.activeClient
	vm.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("no active kafka client")
	}
	slog.Info("create topic", slog.Any("config", config))

	return client.CreateTopic(ctx, config)
}

func (vm *MainViewModel) ProduceMessage(ctx context.Context, topic string, key, value string, headers []models.Header) error {
	vm.mu.RLock()
	client := vm.activeClient
	vm.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("no active kafka client")
	}

	return client.ProduceMessage(ctx, topic, key, value, headers)
}
