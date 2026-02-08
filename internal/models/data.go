package models

import "time"

type Broker struct {
	ID      int
	Name    string
	Address string
}

type Topic struct {
	Name           string
	Partitions     int
	Replicas       int
	InSyncReplicas int
	URP            int
	SegmentSize    int64
	SegmentCount   int
	CleanUpPolicy  string
	MessageCount   int64
	IsInternal     bool
}

type Partition struct {
	ID             int
	MessageCount   int64
	StartOffset    int64
	EndOffset      int64
	Leader         int
	Replicas       []int
	InSyncReplicas []int
}

type ConsumerGroup struct {
	Name    string
	State   string
	Members int
}

type ConsumerGroupOffset struct {
	Topic     string
	Partition int
	Lag       int64
	Offset    int64
}

type Header struct {
	Key   string
	Value string
}

type Message struct {
	Key        string
	Value      string
	Headers    []Header
	Partition  int
	Offset     int64
	Timestamp  time.Time
	Topic      string
}

type MessageFilter struct {
	Partition int
	Offset    int64
	Limit     int
	Format    string
}

type SchemaRegistry struct {
	Subject string
	Version int
	Type    string
	Schema  string
}

func MockBrokers() []Broker {
	return []Broker{}
}

func MockTopics() []Topic {
	return []Topic{
		{Name: "orders", Partitions: 6, Replicas: 3},
		{Name: "payments", Partitions: 3, Replicas: 3},
		{Name: "users", Partitions: 12, Replicas: 3},
		{Name: "notifications", Partitions: 6, Replicas: 2},
		{Name: "analytics-events", Partitions: 24, Replicas: 3},
	}
}

func MockPartitions(topicName string, count int) []Partition {
	partitions := make([]Partition, count)
	for i := range count {
		partitions[i] = Partition{
			ID:           i,
			MessageCount: int64((i + 1) * 500),
			StartOffset:  0,
			EndOffset:    int64((i + 1) * 500),
			Leader:       i % 3,
			Replicas:     []int{0, 1, 2},
		}
	}
	return partitions
}

func MockConsumerGroups() []ConsumerGroup {
	return []ConsumerGroup{
		{Name: "order-processor", State: "Stable", Members: 3},
		{Name: "payment-handler", State: "Stable", Members: 2},
		{Name: "notification-sender", State: "Rebalancing", Members: 4},
		{Name: "analytics-consumer", State: "Stable", Members: 6},
	}
}

func MockConsumerGroupOffsets(groupName string) []ConsumerGroupOffset {
	return []ConsumerGroupOffset{
		{Topic: "orders", Partition: 0, Lag: 0, Offset: 500},
		{Topic: "orders", Partition: 1, Lag: 10, Offset: 490},
		{Topic: "orders", Partition: 2, Lag: 5, Offset: 495},
		{Topic: "payments", Partition: 0, Lag: 0, Offset: 1000},
	}
}

func MockSchemaRegistries() []SchemaRegistry {
	return []SchemaRegistry{
		{
			Subject: "orders-value",
			Version: 3,
			Type:    "AVRO",
			Schema: `{
  "type": "record",
  "name": "Order",
  "namespace": "com.example.orders",
  "fields": [
    {
      "name": "id",
      "type": "string"
    },
    {
      "name": "amount",
      "type": {
        "type": "int",
        "connect.default": 0
      },
      "default": 0
    },
    {
      "name": "status",
      "type": "string"
    }
  ]
}`,
		},
		{
			Subject: "payments-value",
			Version: 2,
			Type:    "AVRO",
			Schema: `{
  "type": "record",
  "name": "Payment",
  "namespace": "com.example.payments",
  "fields": [
    {
      "name": "id",
      "type": "string"
    },
    {
      "name": "orderId",
      "type": "string"
    },
    {
      "name": "amount",
      "type": "double"
    }
  ]
}`,
		},
		{
			Subject: "users-value",
			Version: 5,
			Type:    "JSON",
			Schema: `{
  "type": "object",
  "properties": {
    "id": {"type": "string"},
    "name": {"type": "string"},
    "email": {"type": "string"}
  },
  "required": ["id", "name"]
}`,
		},
	}
}
