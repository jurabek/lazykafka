package viewmodel

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

// MainViewModel coordinates global state and manages child ViewModels
type MainViewModel struct {
	mu sync.RWMutex

	brokersVM                *BrokersViewModel
	topicsVM                 *TopicsViewModel
	consumerGroupsVM         *ConsumerGroupsViewModel
	schemaRegistryVM         *SchemaRegistryViewModel
	topicDetailVM            *TopicDetailViewModel
	consumerGroupDetailVM    *ConsumerGroupDetailViewModel
	schemaRegistryDetailVM   *SchemaRegistryDetailViewModel

	notifyCh chan types.ChangeEvent
	ctx      context.Context
}

// NewMainViewModel creates a new MainViewModel with all child ViewModels
func NewMainViewModel(ctx context.Context) *MainViewModel {
	brokers := models.MockBrokers()
	topics := []models.Topic{}
	consumerGroups := []models.ConsumerGroup{}
	schemaRegistries := []models.SchemaRegistry{}

	vm := &MainViewModel{
		brokersVM:              NewBrokersViewModel(brokers),
		topicsVM:               NewTopicsViewModel(topics),
		consumerGroupsVM:       NewConsumerGroupsViewModel(consumerGroups),
		schemaRegistryVM:       NewSchemaRegistryViewModel(schemaRegistries),
		topicDetailVM:          NewTopicDetailViewModel(),
		consumerGroupDetailVM:  NewConsumerGroupDetailViewModel(),
		schemaRegistryDetailVM: NewSchemaRegistryDetailViewModel(),
		notifyCh:               make(chan types.ChangeEvent),
		ctx:                    ctx,
	}

	vm.startBrokerSubscription()
	vm.setupTopicSelectionCallback()
	vm.setupConsumerGroupSelectionCallback()
	vm.setupSchemaRegistrySelectionCallback()

	// Trigger initial load for the first broker (index 0)
	if broker := vm.brokersVM.GetSelectedBroker(); broker != nil {
		vm.loadDependentData(broker)
	}

	return vm
}

// startBrokerSubscription listens to BrokersViewModel changes and triggers dependent loads
func (vm *MainViewModel) startBrokerSubscription() {
	go func() {
		for {
			select {
			case event := <-vm.brokersVM.NotifyChannel():
				slog.Info("broker vm event changed", slog.Any("field", event.FieldName))
				if event.FieldName == types.FieldSelectedIndex {
					if broker := vm.brokersVM.GetSelectedBroker(); broker != nil {
						vm.loadDependentData(broker)
					}
				}
			case <-vm.ctx.Done():
				return
			}
		}
	}()
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

// loadDependentData triggers async reload of all dependent ViewModels
func (vm *MainViewModel) loadDependentData(broker *models.Broker) {
	vm.topicsVM.LoadForBroker(broker)
	vm.consumerGroupsVM.LoadForBroker(broker)
	vm.schemaRegistryVM.LoadForBroker(broker)
}

func (vm *MainViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *MainViewModel) Notify(fieldName string) {
	vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}
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

func (vm *MainViewModel) ConsumerGroupDetailVM() *ConsumerGroupDetailViewModel {
	return vm.consumerGroupDetailVM
}

func (vm *MainViewModel) SchemaRegistryDetailVM() *SchemaRegistryDetailViewModel {
	return vm.schemaRegistryDetailVM
}
