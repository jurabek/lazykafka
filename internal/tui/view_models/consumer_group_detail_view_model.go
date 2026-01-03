package viewmodel

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type ConsumerGroupDetailViewModel struct {
	mu              sync.RWMutex
	consumerGroup   *models.ConsumerGroup
	offsets         []models.ConsumerGroupOffset
	notifyCh        chan types.ChangeEvent
	commandBindings []*types.CommandBinding
}

func NewConsumerGroupDetailViewModel() *ConsumerGroupDetailViewModel {
	vm := &ConsumerGroupDetailViewModel{
		notifyCh: make(chan types.ChangeEvent),
	}
	vm.commandBindings = []*types.CommandBinding{}
	return vm
}

func (vm *ConsumerGroupDetailViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *ConsumerGroupDetailViewModel) Notify(fieldName string) {
	select {
	case vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}:
	default:
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

func (vm *ConsumerGroupDetailViewModel) SetConsumerGroup(cg *models.ConsumerGroup) {
	vm.mu.Lock()
	vm.consumerGroup = cg
	vm.mu.Unlock()

	if cg != nil {
		offsets := models.MockConsumerGroupOffsets(cg.Name)
		vm.mu.Lock()
		vm.offsets = offsets
		vm.mu.Unlock()
	} else {
		vm.mu.Lock()
		vm.offsets = nil
		vm.mu.Unlock()
	}

	vm.Notify(types.FieldItems)
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

	var sb strings.Builder

	headers := []string{"Topic", "Partition", "Lag", "Offset"}
	colWidths := []int{20, 12, 10, 12}

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
		sb.WriteString(fmt.Sprintf("%-*s%-*d%-*d%-*d\n",
			colWidths[0], o.Topic,
			colWidths[1], o.Partition,
			colWidths[2], o.Lag,
			colWidths[3], o.Offset,
		))
	}

	return sb.String()
}
