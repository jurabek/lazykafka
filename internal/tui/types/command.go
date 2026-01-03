package types

import "fmt"

var ErrNoSelection = fmt.Errorf("no item selected")

type CommandFunc func() error

func NewCommand(f CommandFunc) *Command {
	return &Command{
		notifyCh: make(chan struct{}),
		command:  f,
	}
}

type CommandBinding struct {
	Key      Key
	Name     string
	ViewName string
	Cmd      *Command
}

type Command struct {
	notifyCh chan struct{}
	command  CommandFunc
}

func (c *Command) Execute() error {
	err := c.command()
	if err != nil {
		return err
	}
	c.Notify()
	return nil
}

func (c *Command) Notify() {
	c.notifyCh <- struct{}{}
}

func (c *Command) NotifyChannel() <-chan struct{} {
	return c.notifyCh
}
