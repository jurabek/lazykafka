package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type CGSelectionChangedFunc func(cg *models.ConsumerGroup)

type ConsumerGroupsViewModel struct {
	mu                 sync.RWMutex
	consumerGroups     []models.ConsumerGroup
	selectedIndex      int
	onChange           types.OnChangeFunc
	commandBindings    []*types.CommandBinding
	onSelectionChanged CGSelectionChangedFunc
}

func NewConsumerGroupsViewModel() *ConsumerGroupsViewModel {
	vm := &ConsumerGroupsViewModel{
		selectedIndex: -1,
	}
	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
	}
	return vm
}

func (vm *ConsumerGroupsViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *ConsumerGroupsViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *ConsumerGroupsViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *ConsumerGroupsViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	if index >= 0 && index < len(vm.consumerGroups) {
		vm.selectedIndex = index
		cg := &vm.consumerGroups[index]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(cg)
		}
		return
	}
	vm.mu.Unlock()
}

func (vm *ConsumerGroupsViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.consumerGroups)
}

func (vm *ConsumerGroupsViewModel) MoveUp() error {
	vm.mu.Lock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		cg := &vm.consumerGroups[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(cg)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *ConsumerGroupsViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.consumerGroups)-1 {
		vm.selectedIndex++
		cg := &vm.consumerGroups[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(cg)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *ConsumerGroupsViewModel) SetOnSelectionChanged(fn CGSelectionChangedFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onSelectionChanged = fn
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

func (vm *ConsumerGroupsViewModel) Load(consumerGroups []models.ConsumerGroup) {
	vm.mu.Lock()
	vm.consumerGroups = consumerGroups
	vm.selectedIndex = -1
	vm.mu.Unlock()

	vm.notifyChange(types.FieldItems)
	vm.SetSelectedIndex(0)
}

func (vm *ConsumerGroupsViewModel) LoadForBroker(broker *models.Broker) {
	go func() {
		consumerGroups := models.MockConsumerGroups()
		vm.Load(consumerGroups)
	}()
}
