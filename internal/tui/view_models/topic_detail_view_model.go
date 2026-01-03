package viewmodel

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type TabType int

const (
	TabPartitions TabType = iota
	TabConfiguration
)

type TopicDetailViewModel struct {
	mu              sync.RWMutex
	topic           *models.Topic
	partitions      []models.Partition
	activeTab       TabType
	notifyCh        chan types.ChangeEvent
	commandBindings []*types.CommandBinding
}

func NewTopicDetailViewModel() *TopicDetailViewModel {
	vm := &TopicDetailViewModel{
		activeTab: TabPartitions,
		notifyCh:  make(chan types.ChangeEvent),
	}
	vm.initCommandBindings()
	return vm
}

func (vm *TopicDetailViewModel) initCommandBindings() {
	vm.commandBindings = []*types.CommandBinding{}
}

func (vm *TopicDetailViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *TopicDetailViewModel) Notify(fieldName string) {
	select {
	case vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}:
	default:
	}
}

func (vm *TopicDetailViewModel) GetSelectedIndex() int {
	return 0
}

func (vm *TopicDetailViewModel) SetSelectedIndex(index int) {}

func (vm *TopicDetailViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.partitions)
}

func (vm *TopicDetailViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *TopicDetailViewModel) GetDisplayItems() []string {
	return []string{}
}

func (vm *TopicDetailViewModel) GetTitle() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	if vm.topic != nil {
		return vm.topic.Name
	}
	return "Details"
}

func (vm *TopicDetailViewModel) GetName() string {
	return "topic_detail"
}

func (vm *TopicDetailViewModel) SetTopic(topic *models.Topic) {
	vm.mu.Lock()
	vm.topic = topic
	vm.mu.Unlock()

	if topic != nil {
		partitions := models.MockPartitions(topic.Name, topic.Partitions)
		vm.mu.Lock()
		vm.partitions = partitions
		vm.mu.Unlock()
	} else {
		vm.mu.Lock()
		vm.partitions = nil
		vm.mu.Unlock()
	}

	vm.Notify(types.FieldItems)
}

func (vm *TopicDetailViewModel) GetTopic() *models.Topic {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.topic
}

func (vm *TopicDetailViewModel) GetPartitions() []models.Partition {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.partitions
}

func (vm *TopicDetailViewModel) GetActiveTab() TabType {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.activeTab
}

func (vm *TopicDetailViewModel) SetActiveTab(tab TabType) {
	vm.mu.Lock()
	vm.activeTab = tab
	vm.mu.Unlock()
	vm.Notify(types.FieldSelectedIndex)
}

func (vm *TopicDetailViewModel) RenderPartitionsTable(width int) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.topic == nil {
		return "  Select a topic to view details"
	}

	var sb strings.Builder

	headers := []string{"Partition ID", "Message Count", "Start Offset", "End Offset", "Leader", "Replicas"}
	colWidths := []int{12, 14, 12, 12, 8, 12}

	for i, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s", colWidths[i], h))
	}
	sb.WriteString("\n")

	for i := range headers {
		sb.WriteString(strings.Repeat("-", colWidths[i]-1))
		sb.WriteString(" ")
	}
	sb.WriteString("\n")

	for _, p := range vm.partitions {
		replicas := formatReplicas(p.Replicas)
		sb.WriteString(fmt.Sprintf("%-*d%-*d%-*d%-*d%-*d%-*s\n",
			colWidths[0], p.ID,
			colWidths[1], p.MessageCount,
			colWidths[2], p.StartOffset,
			colWidths[3], p.EndOffset,
			colWidths[4], p.Leader,
			colWidths[5], replicas,
		))
	}

	return sb.String()
}

func formatReplicas(replicas []int) string {
	strs := make([]string, len(replicas))
	for i, r := range replicas {
		strs[i] = fmt.Sprintf("%d", r)
	}
	return strings.Join(strs, ",")
}
