package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const MillsPerDay = 24 * int64(time.Hour/time.Millisecond)

type CleanupPolicy int

const (
	CleanupDelete CleanupPolicy = iota
	CleanupCompact
	CleanupCompactDelete
)

func (c CleanupPolicy) String() string {
	switch c {
	case CleanupCompact:
		return "compact"
	case CleanupCompactDelete:
		return "compact,delete"
	default:
		return "delete"
	}
}

type TopicConfig struct {
	Name              string
	Partitions        int
	ReplicationFactor int
	CleanupPolicy     CleanupPolicy
	MinInSyncReplicas int
	RetentionMs       int64
}

func ParseRetention(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	if daysStr, ok := strings.CutSuffix(s, "d"); ok {
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %w", err)
		}
		return int64(days) * MillsPerDay, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}
	return d.Milliseconds(), nil
}
