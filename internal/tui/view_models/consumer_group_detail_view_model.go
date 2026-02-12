package viewmodel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/kafka"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

const highLagThreshold int64 = 100

type ConsumerGroupDetailViewModel struct {
	mu              sync.RWMutex
	consumerGroup   *models.ConsumerGroup
	offsets         []models.ConsumerGroupOffset
	onChange        types.OnChangeFunc
	commandBindings []*types.CommandBinding
	kafkaClient     kafka.KafkaClient
	onError         func(err error)
}

func NewConsumerGroupDetailViewModel() *ConsumerGroupDetailViewModel {
	vm := &ConsumerGroupDetailViewModel{}
	refresh := types.NewCommand(vm.Refresh)
	vm.commandBindings = []*types.CommandBinding{
		{Key: 'r', Cmd: refresh},
	}
	return vm
}

func (vm *ConsumerGroupDetailViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *ConsumerGroupDetailViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *ConsumerGroupDetailViewModel) GetSelectedIndex() int {
	return 0
}

func (vm *ConsumerGroupDetailViewModel) SetSelectedIndex(index int) {}

func (vm *ConsumerGroupDetailViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.offsets)
}

func (vm *ConsumerGroupDetailViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *ConsumerGroupDetailViewModel) GetDisplayItems() []string {
	return []string{}
}

func (vm *ConsumerGroupDetailViewModel) GetTitle() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	if vm.consumerGroup != nil {
		return vm.consumerGroup.Name
	}
	return "Details"
}

func (vm *ConsumerGroupDetailViewModel) GetName() string {
	return "consumer_group_detail"
}

func (vm *ConsumerGroupDetailViewModel) SetKafkaClient(client kafka.KafkaClient) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kafkaClient = client
}

func (vm *ConsumerGroupDetailViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
}

func (vm *ConsumerGroupDetailViewModel) SetConsumerGroup(cg *models.ConsumerGroup) {
	vm.mu.Lock()
	vm.consumerGroup = cg
	vm.mu.Unlock()

	if cg != nil {
		vm.loadOffsets(cg.Name)
	} else {
		vm.mu.Lock()
		vm.offsets = nil
		vm.mu.Unlock()
	}

	vm.notifyChange(types.FieldItems)
}

func (vm *ConsumerGroupDetailViewModel) loadOffsets(groupName string) {
	vm.mu.RLock()
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.RUnlock()

	if client == nil {
		offsets := models.MockConsumerGroupOffsets(groupName)
		vm.mu.Lock()
		vm.offsets = offsets
		vm.mu.Unlock()
		vm.notifyChange(types.FieldItems)
		return
	}

	go func() {
		offsets, err := client.GetConsumerGroupOffsets(context.Background(), groupName)
		if err != nil {
			slog.Error("failed to load consumer group offsets", slog.Any("error", err))
			if onError != nil {
				onError(err)
			}
			return
		}

		vm.mu.Lock()
		vm.offsets = offsets
		vm.mu.Unlock()
		vm.notifyChange(types.FieldItems)
	}()
}

func (vm *ConsumerGroupDetailViewModel) Refresh() error {
	vm.mu.RLock()
	cg := vm.consumerGroup
	vm.mu.RUnlock()

	if cg == nil {
		return types.ErrNoSelection
	}

	vm.loadOffsets(cg.Name)
	return nil
}

func (vm *ConsumerGroupDetailViewModel) GetConsumerGroup() *models.ConsumerGroup {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.consumerGroup
}

func (vm *ConsumerGroupDetailViewModel) RenderOffsetsTable(width int) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.consumerGroup == nil {
		return "  Select a consumer group to view details"
	}

	if len(vm.offsets) == 0 {
		return "  No offsets available"
	}

	var sb strings.Builder

	headers := []string{"Partition", "Topic", "Current Offset", "End Offset", "Lag"}
	colWidths := []int{12, 20, 16, 12, 10}

	for i, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s", colWidths[i], h))
	}
	sb.WriteString("\n")

	for i := range headers {
		sb.WriteString(strings.Repeat("-", colWidths[i]-1))
		sb.WriteString(" ")
	}
	sb.WriteString("\n")

	for _, o := range vm.offsets {
		row := fmt.Sprintf("%-*d%-*s%-*d%-*d%-*d",
			colWidths[0], o.Partition,
			colWidths[1], o.Topic,
			colWidths[2], o.Offset,
			colWidths[3], o.EndOffset,
			colWidths[4], o.Lag,
		)
		if o.Lag > highLagThreshold {
			sb.WriteString(fmt.Sprintf("\033[31m%s\033[0m\n", row))
		} else {
			sb.WriteString(row + "\n")
		}
	}

	return sb.String()
}
