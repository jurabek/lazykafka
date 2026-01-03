package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type ConsumerGroupsViewModel struct {
	mu             sync.RWMutex
	consumerGroups []models.ConsumerGroup
	selectedIndex  int
	gui            *gocui.Gui
}

func NewConsumerGroupsViewModel(consumerGroups []models.ConsumerGroup, gui *gocui.Gui) *ConsumerGroupsViewModel {
	return &ConsumerGroupsViewModel{
		consumerGroups: consumerGroups,
		gui:            gui,
		selectedIndex:  0,
	}
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

func (vm *ConsumerGroupsViewModel) MoveUp() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return true
	}
	return false
}

func (vm *ConsumerGroupsViewModel) MoveDown() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.consumerGroups)-1 {
		vm.selectedIndex++
		return true
	}
	return false
}

func (vm *ConsumerGroupsViewModel) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
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
	if vm.selectedIndex >= len(consumerGroups) {
		vm.selectedIndex = 0
	}
}

func (vm *ConsumerGroupsViewModel) moveUp() error {
	if vm.MoveUp() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}

func (vm *ConsumerGroupsViewModel) moveDown() error {
	if vm.MoveDown() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}
