package viewmodel

import "github.com/jurabek/lazykafka/internal/tui/types"

type BaseViewModel interface {
	SetOnChange(fn types.OnChangeFunc)
	GetSelectedIndex() int
	SetSelectedIndex(index int)
	GetItemCount() int

	GetCommandBindings() []*types.CommandBinding
	GetDisplayItems() []string
	GetTitle() string
	GetName() string
}
