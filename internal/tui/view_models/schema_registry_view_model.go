package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type SRSelectionChangedFunc func(sr *models.SchemaRegistry)

type SchemaRegistryViewModel struct {
	mu                 sync.RWMutex
	schemaRegistries   []models.SchemaRegistry
	selectedIndex      int
	onChange           types.OnChangeFunc
	commandBindings    []*types.CommandBinding
	onSelectionChanged SRSelectionChangedFunc
}

func NewSchemaRegistryViewModel() *SchemaRegistryViewModel {
	vm := &SchemaRegistryViewModel{
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

func (vm *SchemaRegistryViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *SchemaRegistryViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *SchemaRegistryViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *SchemaRegistryViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	if index >= 0 && index < len(vm.schemaRegistries) {
		vm.selectedIndex = index
		sr := &vm.schemaRegistries[index]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(sr)
		}
		return
	}
	vm.mu.Unlock()
}

func (vm *SchemaRegistryViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.schemaRegistries)
}

func (vm *SchemaRegistryViewModel) MoveUp() error {
	vm.mu.Lock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		sr := &vm.schemaRegistries[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(sr)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *SchemaRegistryViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.schemaRegistries)-1 {
		vm.selectedIndex++
		sr := &vm.schemaRegistries[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(sr)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *SchemaRegistryViewModel) SetOnSelectionChanged(fn SRSelectionChangedFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onSelectionChanged = fn
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

func (vm *SchemaRegistryViewModel) Load(schemaRegistries []models.SchemaRegistry) {
	vm.mu.Lock()
	vm.schemaRegistries = schemaRegistries
	vm.selectedIndex = -1
	vm.mu.Unlock()

	vm.notifyChange(types.FieldItems)
	vm.SetSelectedIndex(0)
}

func (vm *SchemaRegistryViewModel) LoadForBroker(broker *models.Broker) {
	go func() {
		schemaRegistries := models.MockSchemaRegistries()
		vm.Load(schemaRegistries)
	}()
}
