package models

import (
	"testing"
	"time"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "valid message with all fields",
			msg: Message{
				Key:       "test-key",
				Value:     "test-value",
				Headers:   []Header{{Key: "h1", Value: "v1"}},
				Partition: 0,
				Offset:    100,
				Timestamp: time.Now(),
				Topic:     "test-topic",
			},
		},
		{
			name: "message with empty key and value",
			msg: Message{
				Key:       "",
				Value:     "",
				Headers:   nil,
				Partition: 1,
				Offset:    0,
				Timestamp: time.Now(),
				Topic:     "test-topic",
			},
		},
		{
			name: "message with multiple headers",
			msg: Message{
				Key:   "key",
				Value: "value",
				Headers: []Header{
					{Key: "h1", Value: "v1"},
					{Key: "h2", Value: "v2"},
					{Key: "h3", Value: "v3"},
				},
				Partition: 0,
				Offset:    50,
				Timestamp: time.Now(),
				Topic:     "multi-header-topic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Topic == "" {
				t.Error("expected topic to be set")
			}
			if tt.msg.Partition < 0 {
				t.Error("expected partition to be non-negative")
			}
			if tt.msg.Offset < 0 {
				t.Error("expected offset to be non-negative")
			}
		})
	}
}

func TestHeader(t *testing.T) {
	tests := []struct {
		name string
		h    Header
	}{
		{
			name: "valid header",
			h: Header{
				Key:   "content-type",
				Value: "application/json",
			},
		},
		{
			name: "header with empty value",
			h: Header{
				Key:   "empty-header",
				Value: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.h.Key == "" {
				t.Error("expected header key to be set")
			}
		})
	}
}

func TestMessageFilterDefaults(t *testing.T) {
	filter := MessageFilter{
		Partition: -1,
		Offset:    0,
		Limit:     100,
		Format:    "json",
	}

	if filter.Partition != -1 {
		t.Errorf("expected partition -1 for all, got %d", filter.Partition)
	}
	if filter.Offset != 0 {
		t.Errorf("expected offset 0 for newest, got %d", filter.Offset)
	}
	if filter.Limit != 100 {
		t.Errorf("expected limit 100, got %d", filter.Limit)
	}
	if filter.Format != "json" && filter.Format != "plain" {
		t.Errorf("expected format json or plain, got %s", filter.Format)
	}
}

func TestMessageFilterFormats(t *testing.T) {
	validFormats := []string{"json", "plain"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			filter := MessageFilter{Format: format}
			valid := false
			for _, vf := range validFormats {
				if filter.Format == vf {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("format %s is not valid", filter.Format)
			}
		})
	}
}
