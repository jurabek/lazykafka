package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type BrokersViewModel struct {
	mu            sync.RWMutex
	brokers       []models.Broker
	selectedIndex int
	gui           *gocui.Gui
}

func NewBrokersViewModel(brokers []models.Broker, gui *gocui.Gui) *BrokersViewModel {
	return &BrokersViewModel{
		brokers:       brokers,
		gui:           gui,
		selectedIndex: 0,
	}
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

func (vm *BrokersViewModel) MoveUp() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return true
	}
	return false
}

func (vm *BrokersViewModel) MoveDown() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.brokers)-1 {
		vm.selectedIndex++
		return true
	}
	return false
}

func (vm *BrokersViewModel) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
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

func (vm *BrokersViewModel) moveUp() error {
	if vm.MoveUp() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}

func (vm *BrokersViewModel) moveDown() error {
	if vm.MoveDown() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}
