package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type SchemaRegistryDetailViewModel struct {
	mu              sync.RWMutex
	schema          *models.SchemaRegistry
	notifyCh        chan types.ChangeEvent
	commandBindings []*types.CommandBinding
}

func NewSchemaRegistryDetailViewModel() *SchemaRegistryDetailViewModel {
	vm := &SchemaRegistryDetailViewModel{
		notifyCh: make(chan types.ChangeEvent),
	}
	vm.commandBindings = []*types.CommandBinding{}
	return vm
}

func (vm *SchemaRegistryDetailViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *SchemaRegistryDetailViewModel) Notify(fieldName string) {
	select {
	case vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}:
	default:
	}
}

func (vm *SchemaRegistryDetailViewModel) GetSelectedIndex() int {
	return 0
}

func (vm *SchemaRegistryDetailViewModel) SetSelectedIndex(index int) {}

func (vm *SchemaRegistryDetailViewModel) GetItemCount() int {
	return 0
}

func (vm *SchemaRegistryDetailViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *SchemaRegistryDetailViewModel) GetDisplayItems() []string {
	return []string{}
}

func (vm *SchemaRegistryDetailViewModel) GetTitle() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	if vm.schema != nil {
		return fmt.Sprintf("%s (v%d)", vm.schema.Subject, vm.schema.Version)
	}
	return "Schema"
}

func (vm *SchemaRegistryDetailViewModel) GetName() string {
	return "schema_registry_detail"
}

func (vm *SchemaRegistryDetailViewModel) SetSchema(schema *models.SchemaRegistry) {
	vm.mu.Lock()
	vm.schema = schema
	vm.mu.Unlock()
	vm.Notify(types.FieldItems)
}

func (vm *SchemaRegistryDetailViewModel) GetSchema() *models.SchemaRegistry {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.schema
}

func (vm *SchemaRegistryDetailViewModel) GetSchemaContent() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.schema == nil {
		return "  Select a schema to view details"
	}

	return vm.schema.Schema
}

func (vm *SchemaRegistryDetailViewModel) GetSchemaType() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.schema == nil {
		return ""
	}
	return vm.schema.Type
}
