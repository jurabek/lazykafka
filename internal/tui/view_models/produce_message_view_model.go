package viewmodel

import (
	"errors"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

const (
	FieldTopic = 0
	FieldKey   = 1
	FieldValue = 2
)

type ProduceMessageViewModel struct {
	mu           sync.RWMutex
	topic        string
	key          string
	value        string
	headers      map[string]string
	currentField int
	onChange     types.OnChangeFunc
	onSubmit     func(topic string, key, value string, headers []models.Header) error
	onCancel     func()
	validation   map[string]string
}

func NewProduceMessageViewModel(onSubmit func(string, string, string, []models.Header) error, onCancel func()) *ProduceMessageViewModel {
	return &ProduceMessageViewModel{
		headers:      make(map[string]string),
		currentField: FieldKey,
		onSubmit:     onSubmit,
		onCancel:     onCancel,
		validation:   make(map[string]string),
	}
}

func (vm *ProduceMessageViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onChange = fn
}

func (vm *ProduceMessageViewModel) notifyChange(fieldName string) {
	vm.mu.RLock()
	onChange := vm.onChange
	vm.mu.RUnlock()

	if onChange != nil {
		onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *ProduceMessageViewModel) SetTopic(topic string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.topic = topic
	vm.notifyChange("topic")
}

func (vm *ProduceMessageViewModel) GetTopic() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.topic
}

func (vm *ProduceMessageViewModel) SetField(fieldName string, value string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	switch strings.ToLower(fieldName) {
	case "key":
		vm.key = value
	case "value":
		vm.value = value
	case "topic":
		vm.topic = value
	}

	delete(vm.validation, fieldName)
	vm.notifyChange(fieldName)
}

func (vm *ProduceMessageViewModel) GetKey() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.key
}

func (vm *ProduceMessageViewModel) SetKey(key string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.key = key
	delete(vm.validation, "key")
	vm.notifyChange("key")
}

func (vm *ProduceMessageViewModel) GetValue() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.value
}

func (vm *ProduceMessageViewModel) SetValue(value string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.value = value
	delete(vm.validation, "value")
	vm.notifyChange("value")
}

func (vm *ProduceMessageViewModel) SetHeaders(headers map[string]string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.headers = headers
	vm.notifyChange("headers")
}

func (vm *ProduceMessageViewModel) GetHeaders() map[string]string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.headers
}

func (vm *ProduceMessageViewModel) SetHeader(key, value string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.headers == nil {
		vm.headers = make(map[string]string)
	}
	vm.headers[key] = value
	vm.notifyChange("headers")
}

func (vm *ProduceMessageViewModel) GetCurrentField() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentField
}

func (vm *ProduceMessageViewModel) SetCurrentField(field int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.currentField = field
	vm.notifyChange("currentField")
}

func (vm *ProduceMessageViewModel) NextField() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.currentField < FieldValue {
		vm.currentField++
	}
	vm.notifyChange("currentField")
}

func (vm *ProduceMessageViewModel) PrevField() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.currentField > FieldKey {
		vm.currentField--
	}
	vm.notifyChange("currentField")
}

func (vm *ProduceMessageViewModel) GetFieldName(field int) string {
	switch field {
	case FieldTopic:
		return "Topic"
	case FieldKey:
		return "Key"
	case FieldValue:
		return "Value"
	default:
		return ""
	}
}

func (vm *ProduceMessageViewModel) GetFieldValue(field int) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	switch field {
	case FieldTopic:
		return vm.topic
	case FieldKey:
		return vm.key
	case FieldValue:
		return vm.value
	default:
		return ""
	}
}

func (vm *ProduceMessageViewModel) Validate() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.validation = make(map[string]string)

	topic := strings.TrimSpace(vm.topic)
	if topic == "" {
		vm.validation["topic"] = "Topic is required"
		return errors.Join(ErrValidation, errors.New("topic is required"))
	}

	return nil
}

func (vm *ProduceMessageViewModel) GetValidationError(field int) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	fieldName := vm.GetFieldName(field)
	if fieldName == "" {
		return ""
	}
	return vm.validation[strings.ToLower(fieldName)]
}

func (vm *ProduceMessageViewModel) HasValidationErrors() bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.validation) > 0
}

func (vm *ProduceMessageViewModel) Submit() error {
	if err := vm.Validate(); err != nil {
		return err
	}

	vm.mu.RLock()
	topic := strings.TrimSpace(vm.topic)
	key := vm.key
	value := vm.value
	headersMap := vm.headers
	onSubmit := vm.onSubmit
	vm.mu.RUnlock()

	headers := make([]models.Header, 0, len(headersMap))
	for k, v := range headersMap {
		headers = append(headers, models.Header{Key: k, Value: v})
	}

	if onSubmit != nil {
		if err := onSubmit(topic, key, value, headers); err != nil {
			return err
		}
	}
	return nil
}

func (vm *ProduceMessageViewModel) Cancel() {
	vm.mu.RLock()
	onCancel := vm.onCancel
	vm.mu.RUnlock()

	if onCancel != nil {
		onCancel()
	}
}

func (vm *ProduceMessageViewModel) Clear() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.key = ""
	vm.value = ""
	vm.headers = make(map[string]string)
	vm.currentField = FieldKey
	vm.validation = make(map[string]string)
	vm.notifyChange("all")
}

func (vm *ProduceMessageViewModel) GetName() string {
	return "produceMessage"
}

func (vm *ProduceMessageViewModel) GetTitle() string {
	vm.mu.RLock()
	topic := vm.topic
	vm.mu.RUnlock()

	if topic != "" {
		return "Produce Message to: " + topic
	}
	return "Produce Message"
}
