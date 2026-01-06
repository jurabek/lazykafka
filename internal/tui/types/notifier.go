package types

const (
	FieldSelectedIndex = "SelectedIndex"
	FieldItems         = "Items"
)

type ChangeEvent struct {
	FieldName string
}

type OnChangeFunc func(event ChangeEvent)
