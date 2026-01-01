package models

type Broker struct {
	ID   int
	Host string
	Port int
}

type Topic struct {
	Name       string
	Partitions int
	Replicas   int
}

type ConsumerGroup struct {
	Name    string
	State   string
	Members int
}

type SchemaRegistry struct {
	Subject string
	Version int
	Type    string
}

func MockBrokers() []Broker {
	return []Broker{
		{ID: 0, Host: "kafka-broker-0.local", Port: 9092},
		{ID: 1, Host: "kafka-broker-1.local", Port: 9092},
		{ID: 2, Host: "kafka-broker-2.local", Port: 9092},
	}
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

func MockConsumerGroups() []ConsumerGroup {
	return []ConsumerGroup{
		{Name: "order-processor", State: "Stable", Members: 3},
		{Name: "payment-handler", State: "Stable", Members: 2},
		{Name: "notification-sender", State: "Rebalancing", Members: 4},
		{Name: "analytics-consumer", State: "Stable", Members: 6},
	}
}

func MockSchemaRegistries() []SchemaRegistry {
	return []SchemaRegistry{
		{Subject: "orders-value", Version: 3, Type: "AVRO"},
		{Subject: "payments-value", Version: 2, Type: "AVRO"},
		{Subject: "users-value", Version: 5, Type: "JSON"},
	}
}
