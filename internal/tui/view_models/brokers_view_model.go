package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type BrokersViewModel struct {
	mu              sync.RWMutex
	brokers         []models.Broker
	selectedIndex   int
	notifyCh        chan struct{}
	commandBindings []*types.CommandBinding
}

func NewBrokersViewModel(brokers []models.Broker) *BrokersViewModel {
	vm := &BrokersViewModel{
		brokers:       brokers,
		selectedIndex: 0,
		notifyCh:      make(chan struct{}),
	}
	vm.initCommandBindings()
	return vm
}

func (vm *BrokersViewModel) NotifyChannel() <-chan struct{} {
	return vm.notifyCh
}

func (vm *BrokersViewModel) Notify() {
	vm.notifyCh <- struct{}{}
}

func (vm *BrokersViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *BrokersViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if index >= 0 && index < len(vm.brokers) {
		vm.selectedIndex = index
	}
}

func (vm *BrokersViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.brokers)
}

func (vm *BrokersViewModel) MoveUp() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return nil
	}
	return types.ErrNoSelection
}

func (vm *BrokersViewModel) MoveDown() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.brokers)-1 {
		vm.selectedIndex++
		return nil
	}
	return types.ErrNoSelection
}

func (vm *BrokersViewModel) initCommandBindings() {
	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
	}
}

func (vm *BrokersViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *BrokersViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.brokers))
	for i, b := range vm.brokers {
		items[i] = fmt.Sprintf("%d: %s:%d", b.ID, b.Host, b.Port)
	}
	return items
}

func (vm *BrokersViewModel) GetTitle() string {
	return "Brokers"
}

func (vm *BrokersViewModel) GetName() string {
	return "brokers"
}

func (vm *BrokersViewModel) GetSelectedBroker() *models.Broker {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.brokers) {
		return &vm.brokers[vm.selectedIndex]
	}
	return nil
}

func (vm *BrokersViewModel) LoadBrokers(brokers []models.Broker) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.brokers = brokers
	if vm.selectedIndex >= len(brokers) {
		vm.selectedIndex = 0
	}
}
