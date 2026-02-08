package viewmodel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/kafka"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type SelectionChangedFunc func(topic *models.Topic)

type TopicsViewModel struct {
	mu                 sync.RWMutex
	topics             []models.Topic
	selectedIndex      int
	onChange           types.OnChangeFunc
	commandBindings    []*types.CommandBinding
	onSelectionChanged SelectionChangedFunc
	kafkaClient        kafka.KafkaClient
	onError            func(err error)
}

func NewTopicsViewModel() *TopicsViewModel {
	vm := &TopicsViewModel{
		selectedIndex: -1,
	}

	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)
	jumpToBottom := types.NewCommand(vm.JumpToBottom)
	pageDown := types.NewCommand(vm.PageDown)
	pageUp := types.NewCommand(vm.PageUp)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
		{Key: 'G', Cmd: jumpToBottom},
		{Key: gocui.KeyCtrlD, Cmd: pageDown},
		{Key: gocui.KeyCtrlU, Cmd: pageUp},
	}

	return vm
}

func (vm *TopicsViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *TopicsViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *TopicsViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *TopicsViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	if index >= 0 && index < len(vm.topics) {
		vm.selectedIndex = index
		topic := &vm.topics[index]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(topic)
		}
		return
	}
	vm.mu.Unlock()
}

func (vm *TopicsViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.topics)
}

func (vm *TopicsViewModel) MoveUp() error {
	vm.mu.Lock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		topic := &vm.topics[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(topic)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *TopicsViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.topics)-1 {
		vm.selectedIndex++
		topic := &vm.topics[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(topic)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *TopicsViewModel) JumpToTop() error {
	vm.mu.Lock()
	if len(vm.topics) == 0 {
		vm.mu.Unlock()
		return types.ErrNoSelection
	}
	vm.selectedIndex = 0
	topic := &vm.topics[vm.selectedIndex]
	callback := vm.onSelectionChanged
	vm.mu.Unlock()
	if callback != nil {
		callback(topic)
	}
	return nil
}

func (vm *TopicsViewModel) JumpToBottom() error {
	vm.mu.Lock()
	if len(vm.topics) == 0 {
		vm.mu.Unlock()
		return types.ErrNoSelection
	}
	vm.selectedIndex = len(vm.topics) - 1
	topic := &vm.topics[vm.selectedIndex]
	callback := vm.onSelectionChanged
	vm.mu.Unlock()
	if callback != nil {
		callback(topic)
	}
	return nil
}

func (vm *TopicsViewModel) PageDown() error {
	vm.mu.Lock()
	if len(vm.topics) == 0 {
		vm.mu.Unlock()
		return types.ErrNoSelection
	}

	pageSize := 10
	if len(vm.topics) < pageSize {
		pageSize = len(vm.topics)
	}

	newIndex := vm.selectedIndex + pageSize
	if newIndex >= len(vm.topics) {
		newIndex = len(vm.topics) - 1
	}
	vm.selectedIndex = newIndex
	topic := &vm.topics[vm.selectedIndex]
	callback := vm.onSelectionChanged
	vm.mu.Unlock()
	if callback != nil {
		callback(topic)
	}
	return nil
}

func (vm *TopicsViewModel) PageUp() error {
	vm.mu.Lock()
	if len(vm.topics) == 0 {
		vm.mu.Unlock()
		return types.ErrNoSelection
	}

	pageSize := 10
	if len(vm.topics) < pageSize {
		pageSize = len(vm.topics)
	}

	newIndex := vm.selectedIndex - pageSize
	if newIndex < 0 {
		newIndex = 0
	}
	vm.selectedIndex = newIndex
	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.topics) {
		topic := &vm.topics[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		if callback != nil {
			callback(topic)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *TopicsViewModel) SetOnSelectionChanged(fn SelectionChangedFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onSelectionChanged = fn
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

func (vm *TopicsViewModel) Load(topics []models.Topic) {
	vm.mu.Lock()
	vm.topics = topics
	vm.selectedIndex = -1
	vm.mu.Unlock()

	vm.notifyChange(types.FieldItems)
	vm.SetSelectedIndex(0)
}

func (vm *TopicsViewModel) SetKafkaClient(client kafka.KafkaClient) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kafkaClient = client
}

func (vm *TopicsViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
}

func (vm *TopicsViewModel) LoadForBroker(_ *models.Broker) {
	vm.loadTopicsAsync()
}

func (vm *TopicsViewModel) Reload() {
	vm.loadTopicsAsync()
}

func (vm *TopicsViewModel) loadTopicsAsync() {
	vm.mu.RLock()
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.RUnlock()

	if client == nil {
		return
	}

	go func() {
		topics, err := client.ListTopics(context.Background())
		if err != nil {
			slog.Error("failed to load topics", slog.Any("error", err))
			if onError != nil {
				onError(err)
			}
			return
		}
		vm.Load(topics)
	}()
}

func (vm *TopicsViewModel) DeleteTopic(ctx context.Context, topicName string) error {
	vm.mu.RLock()
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("kafka client not configured")
	}

	if err := client.DeleteTopic(ctx, topicName); err != nil {
		slog.Error("failed to delete topic", slog.String("topic", topicName), slog.Any("error", err))
		if onError != nil {
			onError(err)
		}
		return err
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()

	for i, t := range vm.topics {
		if t.Name == topicName {
			vm.topics = append(vm.topics[:i], vm.topics[i+1:]...)
			if vm.selectedIndex >= len(vm.topics) {
				vm.selectedIndex = len(vm.topics) - 1
			}
			break
		}
	}

	vm.notifyChange(types.FieldItems)
	return nil
}
