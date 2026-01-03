package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type ConsumerGroupsViewModel struct {
	mu              sync.RWMutex
	consumerGroups  []models.ConsumerGroup
	selectedIndex   int
	notifyCh        chan types.ChangeEvent
	commandBindings []*types.CommandBinding
}

func NewConsumerGroupsViewModel(consumerGroups []models.ConsumerGroup) *ConsumerGroupsViewModel {
	vm := &ConsumerGroupsViewModel{
		consumerGroups: consumerGroups,
		selectedIndex:  -1,
		notifyCh:       make(chan types.ChangeEvent),
	}
	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)

	commandBindings := []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
	}
	vm.commandBindings = commandBindings
	return vm
}

func (vm *ConsumerGroupsViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *ConsumerGroupsViewModel) Notify(fieldName string) {
	vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}
}

func (vm *ConsumerGroupsViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *ConsumerGroupsViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if index >= 0 && index < len(vm.consumerGroups) {
		vm.selectedIndex = index
	}
}

func (vm *ConsumerGroupsViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.consumerGroups)
}

func (vm *ConsumerGroupsViewModel) MoveUp() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return nil
	}
	return types.ErrNoSelection
}

func (vm *ConsumerGroupsViewModel) MoveDown() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.consumerGroups)-1 {
		vm.selectedIndex++
		return nil
	}
	return types.ErrNoSelection
}

func (vm *ConsumerGroupsViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *ConsumerGroupsViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.consumerGroups))
	for i, cg := range vm.consumerGroups {
		items[i] = fmt.Sprintf("%s [%s] members:%d", cg.Name, cg.State, cg.Members)
	}
	return items
}

func (vm *ConsumerGroupsViewModel) GetTitle() string {
	return "Consumer Groups"
}

func (vm *ConsumerGroupsViewModel) GetName() string {
	return "consumer_groups"
}

func (vm *ConsumerGroupsViewModel) GetSelectedConsumerGroup() *models.ConsumerGroup {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.consumerGroups) {
		return &vm.consumerGroups[vm.selectedIndex]
	}
	return nil
}

func (vm *ConsumerGroupsViewModel) LoadConsumerGroups(consumerGroups []models.ConsumerGroup) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.consumerGroups = consumerGroups
	vm.selectedIndex = -1
}

// LoadForBroker loads consumer groups for the given broker asynchronously
func (vm *ConsumerGroupsViewModel) LoadForBroker(broker *models.Broker) {
	go func() {
		// TODO: Replace with actual Kafka client call to fetch consumer groups for broker
		consumerGroups := models.MockConsumerGroups()

		vm.LoadConsumerGroups(consumerGroups)
		vm.Notify(types.FieldItems)
	}()
}
