package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type SchemaRegistryViewModel struct {
	mu               sync.RWMutex
	schemaRegistries []models.SchemaRegistry
	selectedIndex    int
	notifyCh         chan struct{}
	commandBindings  []*types.CommandBinding
}

func NewSchemaRegistryViewModel(schemaRegistries []models.SchemaRegistry) *SchemaRegistryViewModel {
	vm := &SchemaRegistryViewModel{
		schemaRegistries: schemaRegistries,
		selectedIndex:    0,
		notifyCh:         make(chan struct{}),
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

func (vm *SchemaRegistryViewModel) NotifyChannel() <-chan struct{} {
	return vm.notifyCh
}

func (vm *SchemaRegistryViewModel) Notify() {
	vm.notifyCh <- struct{}{}
}

func (vm *SchemaRegistryViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *SchemaRegistryViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if index >= 0 && index < len(vm.schemaRegistries) {
		vm.selectedIndex = index
	}
}

func (vm *SchemaRegistryViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.schemaRegistries)
}

func (vm *SchemaRegistryViewModel) MoveUp() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return nil
	}
	return types.ErrNoSelection
}

func (vm *SchemaRegistryViewModel) MoveDown() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.schemaRegistries)-1 {
		vm.selectedIndex++
		return nil
	}
	return types.ErrNoSelection
}

func (vm *SchemaRegistryViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *SchemaRegistryViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.schemaRegistries))
	for i, sr := range vm.schemaRegistries {
		items[i] = fmt.Sprintf("%s v%d (%s)", sr.Subject, sr.Version, sr.Type)
	}
	return items
}

func (vm *SchemaRegistryViewModel) GetTitle() string {
	return "Schema Registry"
}

func (vm *SchemaRegistryViewModel) GetName() string {
	return "schema_registry"
}

func (vm *SchemaRegistryViewModel) GetSelectedSchemaRegistry() *models.SchemaRegistry {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.schemaRegistries) {
		return &vm.schemaRegistries[vm.selectedIndex]
	}
	return nil
}

func (vm *SchemaRegistryViewModel) LoadSchemaRegistries(schemaRegistries []models.SchemaRegistry) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.schemaRegistries = schemaRegistries
	if vm.selectedIndex >= len(schemaRegistries) {
		vm.selectedIndex = 0
	}
}
