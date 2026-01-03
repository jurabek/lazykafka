package viewmodel

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type TopicsViewModel struct {
	mu              sync.RWMutex
	topics          []models.Topic
	selectedIndex   int
	notifyCh        chan struct{}
	commandBindings []*types.CommandBinding
}

func NewTopicsViewModel(topics []models.Topic) *TopicsViewModel {
	vm := &TopicsViewModel{
		topics:        topics,
		selectedIndex: 0,
		notifyCh:      make(chan struct{}),
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

func (vm *TopicsViewModel) NotifyChannel() <-chan struct{} {
	return vm.notifyCh
}

func (vm *TopicsViewModel) Notify() {
	vm.notifyCh <- struct{}{}
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

func (vm *TopicsViewModel) MoveUp() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		return nil
	}
	return types.ErrNoSelection
}

func (vm *TopicsViewModel) MoveDown() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.selectedIndex < len(vm.topics)-1 {
		vm.selectedIndex++
		return nil
	}
	return types.ErrNoSelection
}

func (vm *TopicsViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
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
