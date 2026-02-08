# PRD: LazyKafka Full Feature Implementation

## Introduction

LazyKafka is a TUI (Terminal User Interface) application for managing Apache Kafka clusters, built with Go and gocui using MVVM pattern. Currently, it has basic broker and topic listing panels. This PRD defines the implementation of comprehensive message management features to achieve feature parity with kafbat.io's web UI, adapted for a TUI environment.

## Goals

- Enable full message browsing and inspection capabilities
- Support producing messages to topics
- Enable consuming messages with flexible filtering and navigation
- Add partition-level message operations
- Implement topic CRUD operations
- Support basic SASL authentication
- Maintain responsive TUI experience with hybrid (vim/standard) navigation

## User Stories

### US-001: Add Message models to domain layer
**Description:** As a developer, I need message data structures so the application can represent Kafka messages.

**Acceptance Criteria:**
- [ ] Add `Message` struct to `internal/models/data.go` with: Key, Value, Headers, Partition, Offset, Timestamp, Topic
- [ ] Add `MessageFilter` struct with: Partition (int, -1 for all), Offset (int64, 0 for newest), Limit (int, default 100), Format (string: "json"|"plain")
- [ ] Add unit tests for model validation
- [ ] `go test ./internal/models/...` passes

### US-002: Extend KafkaClient interface for message operations
**Description:** As a developer, I need Kafka client methods for producing/consuming messages.

**Acceptance Criteria:**
- [ ] Add `ProduceMessage(ctx, topic, key, value, headers []Header) error` to `KafkaClient` interface
- [ ] Add `ConsumeMessages(ctx, topic, filter MessageFilter) ([]Message, error)` to `KafkaClient` interface
- [ ] Add `DeleteTopic(ctx, topicName string) error` to `KafkaClient` interface
- [ ] Add `GetTopicConfig(ctx, topicName string) (TopicConfig, error)` to `KafkaClient` interface
- [ ] Add `UpdateTopicConfig(ctx, config TopicConfig) error` to `KafkaClient` interface
- [ ] Interface definitions compile without errors

### US-003: Implement message operations in FranzClient
**Description:** As a developer, I need the concrete Kafka client to support message produce/consume.

**Acceptance Criteria:**
- [ ] Implement `ProduceMessage` using franz-go producer
- [ ] Implement `ConsumeMessages` with partition/offset filtering support
- [ ] Support reading from oldest (offset 0) or newest (offset -1)
- [ ] Implement JSON and plain text format handling
- [ ] Handle connection errors gracefully
- [ ] `go build` completes without errors

### US-004: Create MessageBrowserViewModel
**Description:** As a user, I want to browse messages in a topic through a dedicated view model.

**Acceptance Criteria:**
- [ ] Create `internal/tui/view_models/message_browser_view_model.go`
- [ ] Implement message list state management with thread-safe access
- [ ] Support partition selection (all or specific)
- [ ] Support offset mode: oldest-first or newest-first
- [ ] Implement message pagination (load more)
- [ ] Add command bindings: `j`/`k` or arrows for nav, `Enter` to view details, `/` to filter, `r` to refresh
- [ ] `go build` completes without errors

### US-005: Create MessageBrowserView
**Description:** As a user, I want a TUI panel that displays messages in a topic.

**Acceptance Criteria:**
- [ ] Create `internal/tui/views/message_browser_view.go`
- [ ] Display messages in table format: Partition | Offset | Key | Timestamp | Value preview
- [ ] Highlight selected message row
- [ ] Support JSON syntax highlighting for value preview
- [ ] Handle long messages with truncation indicator
- [ ] Show current partition/filter in view title
- [ ] View renders without layout issues

### US-006: Create MessageDetailView for full message inspection
**Description:** As a user, I want to see full message details including headers and full value.

**Acceptance Criteria:**
- [ ] Create `internal/tui/views/message_detail_view.go` with associated view model
- [ ] Display full message metadata: topic, partition, offset, timestamp, headers
- [ ] Display full key and value with JSON pretty-printing when applicable
- [ ] Support scrolling for long messages
- [ ] Add `q` keybinding to close detail view
- [ ] View renders without truncating content

### US-007: Create ProduceMessageViewModel
**Description:** As a user, I want to compose and send messages to a topic.

**Acceptance Criteria:**
- [ ] Create `internal/tui/view_models/produce_message_view_model.go`
- [ ] Manage message composition state: key, value, headers
- [ ] Support JSON validation for value field
- [ ] Handle headers as key-value pairs
- [ ] Validate message before sending
- [ ] Support clearing form
- [ ] `go build` completes without errors

### US-008: Create ProduceMessageView with form input
**Description:** As a user, I want a TUI form to input and send messages.

**Acceptance Criteria:**
- [ ] Create `internal/tui/views/produce_message_view.go`
- [ ] Form fields: Topic (readonly), Key, Value, Headers
- [ ] Tab navigation between fields
- [ ] `Ctrl+S` to send message
- [ ] `Esc` to cancel
- [ ] Show inline validation errors
- [ ] Display success/error status after send attempt

### US-009: Add Messages tab to TopicDetailView
**Description:** As a user, I want to switch between partition info and message browser within topic details.

**Acceptance Criteria:**
- [ ] Extend `TopicDetailViewModel` with `TabMessages` constant
- [ ] Add tab switching logic (e.g., `Tab` key or `1`/`2` keys)
- [ ] Show `MessageBrowserView` when Messages tab active
- [ ] Show existing partitions table when Partitions tab active
- [ ] Update view title to indicate active tab
- [ ] Tab state persists across topic changes

### US-010: Implement Topic CRUD operations
**Description:** As a user, I want to create, view config, and delete topics.

**Acceptance Criteria:**
- [ ] Extend `AddTopicViewModel` to support all `TopicConfig` fields
- [ ] Add `DeleteTopic` method to `TopicsViewModel` with confirmation
- [ ] Add `e` (edit) keybinding on selected topic to show config
- [ ] Add `d` (delete) keybinding with confirmation popup
- [ ] Show topic config popup with editable fields
- [ ] Persist config changes via Kafka client
- [ ] Refresh topic list after CRUD operations

### US-011: Add SASL PLAIN authentication support
**Description:** As a user, I want to connect to Kafka clusters using SASL PLAIN authentication.

**Acceptance Criteria:**
- [ ] Add `Username` and `Password` fields to `BrokerConfig`
- [ ] Add `Mechanism` field to `BrokerConfig` ("plain", "scram-sha-256", "scram-sha-512")
- [ ] Store passwords securely using existing `SecretStore`
- [ ] Pass auth config to Franz client factory
- [ ] Show auth type in broker list display
- [ ] Test connection with auth credentials

### US-012: Implement hybrid keybinding system
**Description:** As a user, I want both vim-style and standard navigation to work.

**Acceptance Criteria:**
- [ ] Support `h`/`j`/`k`/`l` for left/down/up/right navigation
- [ ] Support arrow keys for navigation
- [ ] Add `gg`/`G` to jump to first/last item in lists
- [ ] Add `Ctrl+d`/`Ctrl+u` for page down/up
- [ ] Document all keybindings in help view
- [ ] Keybindings work consistently across all panels

### US-013: Add Consumer Group message inspection
**Description:** As a user, I want to view consumer group offsets and lag per partition.

**Acceptance Criteria:**
- [ ] Extend `ConsumerGroupDetailViewModel` to fetch offsets
- [ ] Add `GetConsumerGroupOffsets(ctx, groupName)` to `KafkaClient`
- [ ] Display partition-level lag in consumer group detail view
- [ ] Show current offset vs end offset for each partition
- [ ] Highlight partitions with significant lag
- [ ] Add `r` keybinding to refresh offset data

### US-014: Add message filtering capabilities
**Description:** As a user, I want to filter messages by partition and offset range.

**Acceptance Criteria:**
- [ ] Add filter popup in message browser (`f` key)
- [ ] Filter options: partition (0-N or all), offset mode (oldest/newest), limit
- [ ] Apply filter and reload messages
- [ ] Show active filter in message browser title
- [ ] Support clearing filters (`Ctrl+f` or dedicated key)
- [ ] Persist filter settings per topic session

### US-015: Add live message tailing
**Description:** As a user, I want to tail new messages in real-time.

**Acceptance Criteria:**
- [ ] Add `t` keybinding to toggle tail mode in message browser
- [ ] In tail mode, poll for new messages every 2 seconds
- [ ] Auto-scroll to newest message when tailing
- [ ] Show "TAILING" indicator in view title
- [ ] Stop tailing on user navigation or view switch
- [ ] Handle tail mode errors gracefully

## Functional Requirements

- FR-1: Application must support producing messages with key, value, and optional headers
- FR-2: Application must support consuming messages from specific partition or all partitions
- FR-3: Application must support reading messages from oldest or newest offset
- FR-4: Message browser must display at least: partition, offset, key, timestamp, value preview
- FR-5: Application must support JSON syntax highlighting in message views
- FR-6: Application must support topic creation, configuration viewing, and deletion
- FR-7: Application must support SASL PLAIN/SCRAM authentication mechanisms
- FR-8: Application must support both vim-style (hjkl) and standard (arrows) keybindings
- FR-9: Application must display consumer group offsets and partition lag
- FR-10: Application must support message filtering by partition, offset mode, and limit
- FR-11: Application must support live tailing of messages
- FR-12: Message values must be displayable as JSON or plain text

## Non-Goals (Out of Scope)

- Metrics dashboard with graphs (TUI limitation)
- OAuth/LDAP authentication (deferred to v2.0)
- SSL/TLS certificate configuration (deferred to v2.0)
- Avro/Protobuf schema integration (deferred to v2.0)
- Schema registry integration (deferred to v2.0)
- Managed Kafka service specific features (Azure EventHub, AWS MSK, etc.)
- Multi-cluster management in single session
- Message search across all topics
- Message export/import functionality

## Design Considerations

### UI/UX
- Maintain existing MVVM architecture
- Reuse existing popup system for forms
- Keep responsive - use goroutines for Kafka I/O
- Show loading indicators during async operations
- Error messages display in help/status bar

### Existing Components to Reuse
- `PopupManager` for forms and confirmations
- `BaseView` for common view functionality
- `CommandBinding` system for keybindings
- `SecretStore` for password storage
- `FranzClientFactory` for Kafka connections

### Navigation Map
- `1-4`: Jump to sidebar panel (Brokers, Topics, CG, Schema)
- `h`/`l` or `←`/`→`: Switch between sidebar panels
- `j`/`k` or `↓`/`↑`: Navigate within list
- `Enter`: View details / Select
- `n`: New (topic/broker)
- `d`: Delete (with confirmation)
- `e`: Edit/config
- `q`: Quit / Close detail
- `Tab`: Switch tabs in detail view
- `r`: Refresh
- `f`: Filter
- `t`: Toggle tail mode
- `Ctrl+S`: Save/Send
- `Esc`: Cancel/Close

## Technical Considerations

- **franz-go library**: Already in use, supports all required operations
- **Context propagation**: Use context for cancellation of long-running operations
- **Thread safety**: ViewModels must use mutex for state changes
- **Error handling**: All Kafka errors should propagate to status bar, not log only
- **Performance**: Limit message fetch to 100-500 messages per request, implement pagination
- **Memory**: Don't keep all messages in memory; implement sliding window or pagination
- **Testing**: Mock `KafkaClient` interface for ViewModel unit tests

## Success Metrics

- User can produce a message in under 10 keystrokes
- Message browser renders 100 messages in under 500ms
- Application remains responsive during Kafka operations
- All keybindings work in both vim and standard modes
- No data races detected with `go test -race`

## Open Questions

1. Should message tailing have configurable polling interval? (Currently hardcoded 2s)
2. What is the maximum message size to display? (Potential memory concern)
3. Should we persist message filter settings across application restarts?
4. Do we need message pagination (load more) or is sliding window sufficient?
