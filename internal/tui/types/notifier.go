package types

// Field name constants for change events
const (
	FieldSelectedIndex = "SelectedIndex"
	FieldItems         = "Items"
)

type ChangeEvent struct {
	FieldName string
}

// Notifier provides change notification via channels
type Notifier interface {
	NotifyChannel() <-chan ChangeEvent
	Notify(fieldName string)
}
