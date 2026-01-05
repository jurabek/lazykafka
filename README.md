# LazyKafka

Terminal UI for Apache Kafka. Browse brokers, topics, consumer groups, and schema registries without writing code.

## Features

- **Broker Management** - Add/manage multiple Kafka clusters with SASL, SSL, AWS IAM auth
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

## Requirements

- Go 1.21+
- Terminal with 60x20 minimum size

## License

MIT
