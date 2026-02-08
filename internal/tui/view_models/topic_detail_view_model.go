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

type TabType int

const (
	TabPartitions TabType = iota
	TabConfiguration
	TabMessages
)

type TopicDetailViewModel struct {
	mu              sync.RWMutex
	topic           *models.Topic
	partitions      []models.Partition
	activeTab       TabType
	onChange        types.OnChangeFunc
	commandBindings []*types.CommandBinding
	kafkaClient     kafka.KafkaClient
	onError         func(err error)
}

func NewTopicDetailViewModel() *TopicDetailViewModel {
	vm := &TopicDetailViewModel{
		activeTab: TabPartitions,
	}
	vm.initCommandBindings()
	return vm
}

func (vm *TopicDetailViewModel) initCommandBindings() {
	vm.commandBindings = []*types.CommandBinding{}
}

func (vm *TopicDetailViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *TopicDetailViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
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

	topicName := "Details"
	if vm.topic != nil {
		topicName = vm.topic.Name
	}

	tabName := ""
	switch vm.activeTab {
	case TabPartitions:
		tabName = "Partitions"
	case TabConfiguration:
		tabName = "Config"
	case TabMessages:
		tabName = "Messages"
	}

	return fmt.Sprintf("%s [%s]", topicName, tabName)
}

func (vm *TopicDetailViewModel) GetName() string {
	return "topic_detail"
}

func (vm *TopicDetailViewModel) SetKafkaClient(client kafka.KafkaClient) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kafkaClient = client
}

func (vm *TopicDetailViewModel) SetOnError(fn func(err error)) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onError = fn
}

func (vm *TopicDetailViewModel) SetTopic(topic *models.Topic) {
	vm.mu.Lock()
	vm.topic = topic
	client := vm.kafkaClient
	onError := vm.onError
	vm.mu.Unlock()

	if topic == nil {
		vm.mu.Lock()
		vm.partitions = nil
		vm.mu.Unlock()
		vm.notifyChange(types.FieldItems)
		return
	}

	if client == nil {
		vm.notifyChange(types.FieldItems)
		return
	}

	go func() {
		partitions, err := client.GetTopicPartitions(context.Background(), topic.Name)
		if err != nil {
			slog.Error("failed to load partitions", slog.Any("error", err))
			if onError != nil {
				onError(err)
			}
			return
		}

		vm.mu.Lock()
		vm.partitions = partitions
		vm.mu.Unlock()
		vm.notifyChange(types.FieldItems)
	}()
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
	vm.notifyChange(types.FieldSelectedIndex)
}

func (vm *TopicDetailViewModel) RenderPartitionsTable(width int) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.topic == nil {
		return "  Select a topic to view details"
	}

	var sb strings.Builder
	t := vm.topic

	topicType := "External"
	if t.IsInternal {
		topicType = "Internal"
	}

	totalReplicas := t.Partitions * t.Replicas
	isrDisplay := fmt.Sprintf("%d of %d", t.InSyncReplicas, totalReplicas)

	sb.WriteString(fmt.Sprintf("%-20s%-20s%-20s%-20s\n", "Partitions", "Replication Factor", "URP", "In Sync Replicas"))
	sb.WriteString(fmt.Sprintf("%-20d%-20d%-20d%-20s\n\n", t.Partitions, t.Replicas, t.URP, isrDisplay))

	sb.WriteString(fmt.Sprintf("%-20s%-20s%-20s\n", "Type", "Clean Up Policy", "Message Count"))
	sb.WriteString(fmt.Sprintf("%-20s%-20s%-20d\n\n", topicType, t.CleanUpPolicy, t.MessageCount))

	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n\n")

	headers := []string{"Partition", "Replicas", "First Offset", "Next Offset", "Message Count"}
	colWidths := []int{12, 12, 14, 14, 14}

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
		sb.WriteString(fmt.Sprintf("%-*d%-*s%-*d%-*d%-*d\n",
			colWidths[0], p.ID,
			colWidths[1], replicas,
			colWidths[2], p.StartOffset,
			colWidths[3], p.EndOffset,
			colWidths[4], p.MessageCount,
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
