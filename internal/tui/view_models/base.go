package viewmodel

import "github.com/jurabek/lazykafka/internal/tui/types"

type BaseViewModel interface {
	GetSelectedIndex() int
	SetSelectedIndex(index int)
	GetItemCount() int

	MoveUp() bool
	MoveDown() bool

	GetKeybindings(opts types.KeybindingsOpts) []*types.Binding

	GetDisplayItems() []string
	GetTitle() string
	GetName() string
}
