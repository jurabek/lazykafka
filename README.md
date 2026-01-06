# LazyKafka

Terminal UI for Apache Kafka. Browse brokers, topics, consumer groups, and schema registries without writing code.

## Features

- **Broker Management** - Add/manage multiple Kafka clusters with SASL authentication or no authentication
- **Topics** - Browse topics, view partitions and metadata
- **Consumer Groups** - Monitor groups, offsets, lag, and members
- **Schema Registry** - View schemas, versions, and definitions

## Install

```bash
go install github.com/jurabek/lazykafka/cmd/lazykafka@latest
```

Or build from source:

```bash
git clone https://github.com/jurabek/lazykafka.git
cd lazykafka
go build -o lazykafka ./cmd/lazykafka
```

## Usage

```bash
./lazykafka
```

## Keybindings

## Configuration

Broker configs stored in `~/.lazykafka/brokers.json`

## Authentication

LazyKafka supports the following authentication methods for connecting to Kafka brokers:

### No Authentication
- **Auth Type**: 0
- **Required Fields**: None

### SASL Authentication
- **Auth Type**: 1
- **Required Fields**:
  - `sasl_mechanism`: 0=PLAIN, 1=SCRAM-SHA-256, 2=SCRAM-SHA-512, 3=OAUTHBEARER
  - `username`: SASL username
  - `password`: SASL password

Example broker configuration:
```json
{
  "name": "secure-cluster",
  "bootstrap_servers": "kafka.example.com:9092",
  "auth_type": 1,
  "sasl_mechanism": 1,
  "username": "myuser",
}
```

## Requirements

- Go 1.21+
- Terminal with 60x20 minimum size

## License

