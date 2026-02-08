package viewmodel

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type MessageDetailViewModel struct {
	mu              sync.RWMutex
	currentMessage  *models.Message
	scrollPosition  int
	onChange        types.OnChangeFunc
	onClose         func()
	commandBindings []*types.CommandBinding
	displayLines    []string
}

func NewMessageDetailViewModel() *MessageDetailViewModel {
	vm := &MessageDetailViewModel{
		currentMessage: nil,
		scrollPosition: 0,
		displayLines:   []string{},
	}

	scrollUp := types.NewCommand(vm.ScrollUp)
	scrollDown := types.NewCommand(vm.ScrollDown)
	close := types.NewCommand(vm.Close)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: scrollUp},
		{Key: 'j', Cmd: scrollDown},
		{Key: gocui.KeyArrowUp, Cmd: scrollUp},
		{Key: gocui.KeyArrowDown, Cmd: scrollDown},
		{Key: gocui.KeyPgup, Cmd: types.NewCommand(func() error {
			for i := 0; i < 5; i++ {
				vm.ScrollUp()
			}
			return nil
		})},
		{Key: gocui.KeyPgdn, Cmd: types.NewCommand(func() error {
			for i := 0; i < 5; i++ {
				vm.ScrollDown()
			}
			return nil
		})},
		{Key: 'q', Cmd: close},
		{Key: gocui.KeyEsc, Cmd: close},
	}

	return vm
}

func (vm *MessageDetailViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *MessageDetailViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *MessageDetailViewModel) GetSelectedIndex() int {
	return 0
}

func (vm *MessageDetailViewModel) SetSelectedIndex(index int) {}

func (vm *MessageDetailViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.displayLines)
}

func (vm *MessageDetailViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *MessageDetailViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.displayLines
}

func (vm *MessageDetailViewModel) GetTitle() string {
	vm.mu.RLock()
	msg := vm.currentMessage
	vm.mu.RUnlock()

	if msg != nil && msg.Topic != "" {
		return fmt.Sprintf("Message: %s", msg.Topic)
	}
	return "Message Details"
}

func (vm *MessageDetailViewModel) GetName() string {
	return "message_detail"
}

func (vm *MessageDetailViewModel) SetOnClose(fn func()) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onClose = fn
}

func (vm *MessageDetailViewModel) Close() error {
	vm.mu.RLock()
	onClose := vm.onClose
	vm.mu.RUnlock()

	if onClose != nil {
		onClose()
	}
	return nil
}

func (vm *MessageDetailViewModel) SetMessage(msg *models.Message) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.currentMessage = msg
	vm.scrollPosition = 0
	vm.displayLines = vm.buildDisplayText(msg)
	vm.notifyChange(types.FieldItems)
}

func (vm *MessageDetailViewModel) GetMessage() *models.Message {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentMessage
}

func (vm *MessageDetailViewModel) ScrollUp() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.scrollPosition > 0 {
		vm.scrollPosition--
	}
	return nil
}

func (vm *MessageDetailViewModel) ScrollDown() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	maxScroll := len(vm.displayLines) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	if vm.scrollPosition < maxScroll {
		vm.scrollPosition++
	}
	return nil
}

func (vm *MessageDetailViewModel) GetScrollPosition() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.scrollPosition
}

func (vm *MessageDetailViewModel) GetDisplayText() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.displayLines
}

func (vm *MessageDetailViewModel) buildDisplayText(msg *models.Message) []string {
	if msg == nil {
		return []string{"No message selected"}
	}

	var lines []string
	separator := strings.Repeat("-", 60)

	lines = append(lines, separator)
	lines = append(lines, "METADATA")
	lines = append(lines, separator)
	lines = append(lines, fmt.Sprintf("  Topic:      %s", msg.Topic))
	lines = append(lines, fmt.Sprintf("  Partition:  %d", msg.Partition))
	lines = append(lines, fmt.Sprintf("  Offset:     %d", msg.Offset))
	lines = append(lines, fmt.Sprintf("  Timestamp:  %s", msg.Timestamp.Format("2006-01-02 15:04:05.000")))
	lines = append(lines, "")

	if len(msg.Headers) > 0 {
		lines = append(lines, separator)
		lines = append(lines, "HEADERS")
		lines = append(lines, separator)
		for _, h := range msg.Headers {
			lines = append(lines, fmt.Sprintf("  %s: %s", h.Key, h.Value))
		}
		lines = append(lines, "")
	}

	lines = append(lines, separator)
	lines = append(lines, "KEY")
	lines = append(lines, separator)
	if msg.Key != "" {
		lines = append(lines, fmt.Sprintf("  %s", msg.Key))
	} else {
		lines = append(lines, "  (null)")
	}
	lines = append(lines, "")

	lines = append(lines, separator)
	lines = append(lines, "VALUE")
	lines = append(lines, separator)

	if msg.Value != "" {
		if vm.isJSON(msg.Value) {
			var prettyJSON interface{}
			if err := json.Unmarshal([]byte(msg.Value), &prettyJSON); err == nil {
				formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
				valueLines := strings.Split(string(formatted), "\n")
				for _, line := range valueLines {
					lines = append(lines, "  "+line)
				}
			} else {
				for _, line := range strings.Split(msg.Value, "\n") {
					lines = append(lines, "  "+line)
				}
			}
		} else {
			for _, line := range strings.Split(msg.Value, "\n") {
				lines = append(lines, "  "+line)
			}
		}
	} else {
		lines = append(lines, "  (null)")
	}

	return lines
}

func (vm *MessageDetailViewModel) isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}
