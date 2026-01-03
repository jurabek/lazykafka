package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type TopicsViewModel struct {
	mu            sync.RWMutex
	topics        []models.Topic
	selectedIndex int
	gui           *gocui.Gui
}

func NewTopicsViewModel(topics []models.Topic, gui *gocui.Gui) *TopicsViewModel {
	return &TopicsViewModel{
		topics:        topics,
		gui:           gui,
		selectedIndex: 0,
	}
}

func (vm *TopicsViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *TopicsViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if index >= 0 && index < len(vm.topics) {
		vm.selectedIndex = index
	}
}

func (vm *TopicsViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.topics)
}

func (vm *TopicsViewModel) MoveUp() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return true
	}
	return false
}

func (vm *TopicsViewModel) MoveDown() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.topics)-1 {
		vm.selectedIndex++
		return true
	}
	return false
}

func (vm *TopicsViewModel) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
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

func (vm *TopicsViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.topics))
	for i, t := range vm.topics {
		items[i] = fmt.Sprintf("%s (P:%d R:%d)", t.Name, t.Partitions, t.Replicas)
	}
	return items
}

func (vm *TopicsViewModel) GetTitle() string {
	return "Topics"
}

func (vm *TopicsViewModel) GetName() string {
	return "topics"
}

func (vm *TopicsViewModel) GetSelectedTopic() *models.Topic {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.topics) {
		return &vm.topics[vm.selectedIndex]
	}
	return nil
}

func (vm *TopicsViewModel) LoadTopics(topics []models.Topic) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.topics = topics
	if vm.selectedIndex >= len(topics) {
		vm.selectedIndex = 0
	}
}

func (vm *TopicsViewModel) moveUp() error {
	if vm.MoveUp() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}

func (vm *TopicsViewModel) moveDown() error {
	if vm.MoveDown() {
		vm.gui.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
	return nil
}
