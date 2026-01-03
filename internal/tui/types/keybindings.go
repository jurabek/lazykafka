package types

import "github.com/jroimartin/gocui"

type Key interface{}

type KeybindingsOpts struct {
	GetKey func(key string) Key
	Config KeybindingConfig
}

type KeybindingConfig struct {
}

type Binding struct {
	ViewName    string
	Key         Key
	Modifier    gocui.Modifier
	Handler     func() error
	Description string
	Tooltip     string
}
