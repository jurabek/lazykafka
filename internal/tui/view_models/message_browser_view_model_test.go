package viewmodel

import (
	"context"
	"testing"

	"github.com/jurabek/lazykafka/internal/models"
)

type mockKafkaClient struct {
	consumeMessages    []models.Message
	consumeErr         error
	consumeCalledWith  models.MessageFilter
	consumeCalledTopic string
}

func (m *mockKafkaClient) Connect(_ context.Context) error { return nil }
func (m *mockKafkaClient) Close()                          {}
func (m *mockKafkaClient) ListTopics(_ context.Context) ([]models.Topic, error) {
	return nil, nil
}
func (m *mockKafkaClient) GetTopicPartitions(_ context.Context, _ string) ([]models.Partition, error) {
	return nil, nil
}
func (m *mockKafkaClient) CreateTopic(_ context.Context, _ models.TopicConfig) error { return nil }
func (m *mockKafkaClient) ProduceMessage(_ context.Context, _ string, _ string, _ string, _ []models.Header) error {
	return nil
}
func (m *mockKafkaClient) ConsumeMessages(_ context.Context, topic string, filter models.MessageFilter) ([]models.Message, error) {
	m.consumeCalledWith = filter
	m.consumeCalledTopic = topic
	return m.consumeMessages, m.consumeErr
}
func (m *mockKafkaClient) DeleteTopic(_ context.Context, _ string) error { return nil }
func (m *mockKafkaClient) GetTopicConfig(_ context.Context, _ string) (models.TopicConfig, error) {
	return models.TopicConfig{}, nil
}
func (m *mockKafkaClient) UpdateTopicConfig(_ context.Context, _ models.TopicConfig) error {
	return nil
}
func (m *mockKafkaClient) GetConsumerGroupOffsets(_ context.Context, _ string) ([]models.ConsumerGroupOffset, error) {
	return nil, nil
}

func TestShowFilterPopup(t *testing.T) {
	tests := []struct {
		name           string
		currentFilter  models.MessageFilter
		wantPartition  int
		wantOffsetMode string
		wantLimit      int
		wantPopupOpen  bool
	}{
		{
			name: "opens popup with current filter values (newest)",
			currentFilter: models.MessageFilter{
				Partition: -1,
				Offset:    -1,
				Limit:     100,
				Format:    "json",
			},
			wantPartition:  -1,
			wantOffsetMode: OffsetModeNewest,
			wantLimit:      100,
			wantPopupOpen:  true,
		},
		{
			name: "opens popup with current filter values (oldest)",
			currentFilter: models.MessageFilter{
				Partition: 2,
				Offset:    0,
				Limit:     50,
				Format:    "json",
			},
			wantPartition:  2,
			wantOffsetMode: OffsetModeOldest,
			wantLimit:      50,
			wantPopupOpen:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			vm := NewMessageBrowserViewModel()
			vm.currentFilter = tt.currentFilter

			vm.ShowFilterPopup()

			if got := vm.IsFilterPopupOpen(); got != tt.wantPopupOpen {
				t.Errorf("IsFilterPopupOpen() = %v, want %v", got, tt.wantPopupOpen)
			}
			if got := vm.GetPendingPartition(); got != tt.wantPartition {
				t.Errorf("GetPendingPartition() = %v, want %v", got, tt.wantPartition)
			}
			if got := vm.GetPendingOffsetMode(); got != tt.wantOffsetMode {
				t.Errorf("GetPendingOffsetMode() = %v, want %v", got, tt.wantOffsetMode)
			}
			if got := vm.GetPendingLimit(); got != tt.wantLimit {
				t.Errorf("GetPendingLimit() = %v, want %v", got, tt.wantLimit)
			}
		})
	}
}

func TestApplyFilter(t *testing.T) {
	tests := []struct {
		name              string
		pendingPartition  int
		pendingOffsetMode string
		pendingLimit      int
		wantFilter        models.MessageFilter
	}{
		{
			name:              "apply newest offset mode",
			pendingPartition:  -1,
			pendingOffsetMode: OffsetModeNewest,
			pendingLimit:      100,
			wantFilter: models.MessageFilter{
				Partition: -1,
				Offset:    -1,
				Limit:     100,
				Format:    "json",
			},
		},
		{
			name:              "apply oldest offset mode with specific partition",
			pendingPartition:  3,
			pendingOffsetMode: OffsetModeOldest,
			pendingLimit:      50,
			wantFilter: models.MessageFilter{
				Partition: 3,
				Offset:    0,
				Limit:     50,
				Format:    "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			vm := NewMessageBrowserViewModel()
			mock := &mockKafkaClient{}
			vm.SetKafkaClient(mock)
			vm.SetTopic("test-topic")

			vm.ShowFilterPopup()
			vm.SetPendingPartition(tt.pendingPartition)
			vm.SetPendingOffsetMode(tt.pendingOffsetMode)
			vm.SetPendingLimit(tt.pendingLimit)

			vm.ApplyFilter()

			if vm.IsFilterPopupOpen() {
				t.Error("expected filter popup to be closed after apply")
			}

			got := vm.GetFilter()
			if got.Partition != tt.wantFilter.Partition {
				t.Errorf("filter.Partition = %v, want %v", got.Partition, tt.wantFilter.Partition)
			}
			if got.Offset != tt.wantFilter.Offset {
				t.Errorf("filter.Offset = %v, want %v", got.Offset, tt.wantFilter.Offset)
			}
			if got.Limit != tt.wantFilter.Limit {
				t.Errorf("filter.Limit = %v, want %v", got.Limit, tt.wantFilter.Limit)
			}
		})
	}
}

func TestClearFilter(t *testing.T) {
	t.Parallel()

	vm := NewMessageBrowserViewModel()
	mock := &mockKafkaClient{}
	vm.SetKafkaClient(mock)
	vm.SetTopic("test-topic")

	vm.ShowFilterPopup()
	vm.SetPendingPartition(5)
	vm.SetPendingOffsetMode(OffsetModeOldest)
	vm.SetPendingLimit(25)

	vm.ClearFilter()

	if vm.IsFilterPopupOpen() {
		t.Error("expected filter popup to be closed after clear")
	}

	got := vm.GetFilter()
	if got.Partition != -1 {
		t.Errorf("filter.Partition = %v, want -1", got.Partition)
	}
	if got.Offset != -1 {
		t.Errorf("filter.Offset = %v, want -1", got.Offset)
	}
	if got.Limit != 100 {
		t.Errorf("filter.Limit = %v, want 100", got.Limit)
	}
}

func TestCloseFilterPopup(t *testing.T) {
	t.Parallel()

	vm := NewMessageBrowserViewModel()
	vm.ShowFilterPopup()

	if !vm.IsFilterPopupOpen() {
		t.Error("expected filter popup to be open")
	}

	vm.CloseFilterPopup()

	if vm.IsFilterPopupOpen() {
		t.Error("expected filter popup to be closed")
	}
}

func TestPendingFilterSettersGetters(t *testing.T) {
	tests := []struct {
		name      string
		partition int
		mode      string
		limit     int
	}{
		{
			name:      "default values",
			partition: -1,
			mode:      OffsetModeNewest,
			limit:     100,
		},
		{
			name:      "specific partition oldest",
			partition: 2,
			mode:      OffsetModeOldest,
			limit:     50,
		},
		{
			name:      "all partitions with small limit",
			partition: -1,
			mode:      OffsetModeNewest,
			limit:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			vm := NewMessageBrowserViewModel()

			vm.SetPendingPartition(tt.partition)
			vm.SetPendingOffsetMode(tt.mode)
			vm.SetPendingLimit(tt.limit)

			if got := vm.GetPendingPartition(); got != tt.partition {
				t.Errorf("GetPendingPartition() = %v, want %v", got, tt.partition)
			}
			if got := vm.GetPendingOffsetMode(); got != tt.mode {
				t.Errorf("GetPendingOffsetMode() = %v, want %v", got, tt.mode)
			}
			if got := vm.GetPendingLimit(); got != tt.limit {
				t.Errorf("GetPendingLimit() = %v, want %v", got, tt.limit)
			}
		})
	}
}

func TestOffsetModeConversion(t *testing.T) {
	tests := []struct {
		name       string
		offset     int64
		wantMode   string
		mode       string
		wantOffset int64
	}{
		{
			name:       "newest offset converts to newest mode",
			offset:     -1,
			wantMode:   OffsetModeNewest,
			mode:       OffsetModeNewest,
			wantOffset: -1,
		},
		{
			name:       "oldest offset converts to oldest mode",
			offset:     0,
			wantMode:   OffsetModeOldest,
			mode:       OffsetModeOldest,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := offsetToMode(tt.offset); got != tt.wantMode {
				t.Errorf("offsetToMode(%d) = %v, want %v", tt.offset, got, tt.wantMode)
			}
			if got := modeToOffset(tt.mode); got != tt.wantOffset {
				t.Errorf("modeToOffset(%s) = %v, want %v", tt.mode, got, tt.wantOffset)
			}
		})
	}
}

func TestNewMessageBrowserViewModelDefaults(t *testing.T) {
	t.Parallel()

	vm := NewMessageBrowserViewModel()

	if vm.selectedIndex != -1 {
		t.Errorf("selectedIndex = %v, want -1", vm.selectedIndex)
	}
	if vm.pendingPartition != -1 {
		t.Errorf("pendingPartition = %v, want -1", vm.pendingPartition)
	}
	if vm.pendingOffsetMode != OffsetModeNewest {
		t.Errorf("pendingOffsetMode = %v, want %v", vm.pendingOffsetMode, OffsetModeNewest)
	}
	if vm.pendingLimit != 100 {
		t.Errorf("pendingLimit = %v, want 100", vm.pendingLimit)
	}
	if vm.filterPopupOpen {
		t.Error("filterPopupOpen should be false by default")
	}
}
