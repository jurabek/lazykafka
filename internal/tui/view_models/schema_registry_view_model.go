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
	gui              *gocui.Gui
}

func NewSchemaRegistryViewModel(schemaRegistries []models.SchemaRegistry, gui *gocui.Gui) *SchemaRegistryViewModel {
	return &SchemaRegistryViewModel{
		schemaRegistries: schemaRegistries,
		gui:              gui,
		selectedIndex:    0,
	}
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

func (vm *SchemaRegistryViewModel) MoveUp() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return true
	}
	return false
}

func (vm *SchemaRegistryViewModel) MoveDown() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.schemaRegistries)-1 {
		vm.selectedIndex++
		return true
	}
	return false
}

func (vm *SchemaRegistryViewModel) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			ViewName:    vm.GetName(),
			Key:         'k',
			Modifier:    gocui.ModNone,
			Handler:     vm.moveUp,
			Description: "move up",
		},
		{
			ViewName:    vm.GetName(),
			Key:         'j',
			Modifier:    gocui.ModNone,
			Handler:     vm.moveDown,
			Description: "move down",
		},
		{
			ViewName:    vm.GetName(),
			Key:         gocui.KeyArrowUp,
			Modifier:    gocui.ModNone,
			Handler:     vm.moveUp,
			Description: "move up",
		},
		{
			ViewName:    vm.GetName(),
			Key:         gocui.KeyArrowDown,
			Modifier:    gocui.ModNone,
			Handler:     vm.moveDown,
			Description: "move down",
		},
	}
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

func (vm *SchemaRegistryViewModel) moveUp() error {
	if vm.MoveUp() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}

func (vm *SchemaRegistryViewModel) moveDown() error {
	if vm.MoveDown() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}
