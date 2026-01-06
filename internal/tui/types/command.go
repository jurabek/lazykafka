package types

import "fmt"

var ErrNoSelection = fmt.Errorf("no item selected")
var ErrRestartApp = fmt.Errorf("restart app")

type CommandFunc func() error

func NewCommand(f CommandFunc) *Command {
	return &Command{command: f}
}

type CommandBinding struct {
	Key      Key
	Name     string
	ViewName string
	Cmd      *Command
}

type Command struct {
	command    CommandFunc
	onExecuted func()
}

func (c *Command) SetOnExecuted(fn func()) {
	c.onExecuted = fn
}

func (c *Command) Execute() error {
	if err := c.command(); err != nil {
		return err
	}
	if c.onExecuted != nil {
		c.onExecuted()
	}
	return nil
}
