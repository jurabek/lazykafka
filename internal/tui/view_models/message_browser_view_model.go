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

type MessageSelectedFunc func(msg *models.Message)

type MessageBrowserViewModel struct {
	mu                sync.RWMutex
	messages          []models.Message
	selectedIndex     int
	currentFilter     models.MessageFilter
	currentTopic      string
	onChange          types.OnChangeFunc
	commandBindings   []*types.CommandBinding
	onMessageSelected MessageSelectedFunc
	kafkaClient       kafka.KafkaClient
	onError           func(err error)
}

func NewMessageBrowserViewModel() *MessageBrowserViewModel {
	vm := &MessageBrowserViewModel{
		selectedIndex: -1,
		currentFilter: models.MessageFilter{
			Partition: -1,
			Offset:    -1,
			Limit:     100,
			Format:    "json",
		},
	}

	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)
	refresh := types.NewCommand(vm.Refresh)
	selectMsg := types.NewCommand(vm.SelectMessage)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
		{Key: 'r', Cmd: refresh},
		{Key: gocui.KeyEnter, Cmd: selectMsg},
	}

	return vm
}

func (vm *MessageBrowserViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *MessageBrowserViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *MessageBrowserViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *MessageBrowserViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if index >= 0 && index < len(vm.messages) {
		vm.selectedIndex = index
		if vm.onMessageSelected != nil {
			msg := &vm.messages[index]
			vm.onMessageSelected(msg)
		}
	}
}

func (vm *MessageBrowserViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.messages)
}

func (vm *MessageBrowserViewModel) MoveUp() error {
	vm.mu.Lock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		msg := &vm.messages[vm.selectedIndex]
		callback := vm.onMessageSelected
		vm.mu.Unlock()
		if callback != nil {
			callback(msg)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *MessageBrowserViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.messages)-1 {
		vm.selectedIndex++
		msg := &vm.messages[vm.selectedIndex]
		callback := vm.onMessageSelected
		vm.mu.Unlock()
		if callback != nil {
			callback(msg)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *MessageBrowserViewModel) SetOnMessageSelected(fn MessageSelectedFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onMessageSelected = fn
}

func (vm *MessageBrowserViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *MessageBrowserViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.messages))
	for i, m := range vm.messages {
		valuePreview := m.Value
		if len(valuePreview) > 30 {
			valuePreview = valuePreview[:30] + "..."
		}
		items[i] = fmt.Sprintf("P:%d O:%d | K:%s | V:%s | %s",
			m.Partition, m.Offset, m.Key, valuePreview, m.Timestamp.Format("15:04:05"))
	}
	return items
}

func (vm *MessageBrowserViewModel) GetTitle() string {
	vm.mu.RLock()
	topic := vm.currentTopic
	filter := vm.currentFilter
	vm.mu.RUnlock()

	if topic == "" {
		return "Messages"
	}
	return fmt.Sprintf("%s [P:%d L:%d]", topic, filter.Partition, filter.Limit)
}

func (vm *MessageBrowserViewModel) GetName() string {
	return "messages"
}

func (vm *MessageBrowserViewModel) GetSelectedMessage() *models.Message {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.messages) {
		return &vm.messages[vm.selectedIndex]
	}
	return nil
}

func (vm *MessageBrowserViewModel) SetTopic(topic string) {
	vm.mu.Lock()
	vm.currentTopic = topic
	vm.mu.Unlock()
	vm.notifyChange(types.FieldItems)
}

func (vm *MessageBrowserViewModel) LoadMessages(filter models.MessageFilter) {
	vm.mu.Lock()
	vm.currentFilter = filter
	topic := vm.currentTopic
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.Unlock()

	if client == nil || topic == "" {
		return
	}

	go func() {
		messages, err := client.ConsumeMessages(context.Background(), topic, filter)
		if err != nil {
			slog.Error("failed to load messages", slog.Any("error", err))
			if onError != nil {
				onError(err)
			}
			return
		}

		vm.mu.Lock()
		vm.messages = messages
		vm.selectedIndex = -1
		vm.mu.Unlock()

		vm.notifyChange(types.FieldItems)
		if len(messages) > 0 {
			vm.SetSelectedIndex(0)
		}
	}()
}

func (vm *MessageBrowserViewModel) Refresh() error {
	vm.mu.RLock()
	filter := vm.currentFilter
	vm.mu.RUnlock()
	vm.LoadMessages(filter)
	return nil
}

func (vm *MessageBrowserViewModel) SelectMessage() error {
	vm.mu.RLock()
	selectedIndex := vm.selectedIndex
	messages := vm.messages
	callback := vm.onMessageSelected
	vm.mu.RUnlock()

	if selectedIndex < 0 || selectedIndex >= len(messages) {
		return types.ErrNoSelection
	}

	if callback != nil {
		callback(&messages[selectedIndex])
	}
	return nil
}

func (vm *MessageBrowserViewModel) SetKafkaClient(client kafka.KafkaClient) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kafkaClient = client
}

func (vm *MessageBrowserViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
}

func (vm *MessageBrowserViewModel) GetFilter() models.MessageFilter {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentFilter
}
