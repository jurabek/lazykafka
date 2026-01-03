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

	brokersVM        *BrokersViewModel
	topicsVM         *TopicsViewModel
	consumerGroupsVM *ConsumerGroupsViewModel
	schemaRegistryVM *SchemaRegistryViewModel

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
		brokersVM:        NewBrokersViewModel(brokers),
		topicsVM:         NewTopicsViewModel(topics),
		consumerGroupsVM: NewConsumerGroupsViewModel(consumerGroups),
		schemaRegistryVM: NewSchemaRegistryViewModel(schemaRegistries),
		notifyCh:         make(chan types.ChangeEvent),
		ctx:              ctx,
	}

	vm.startBrokerSubscription()

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
