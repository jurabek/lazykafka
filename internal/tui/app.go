package tui

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type App struct {
	gui               *gocui.Gui
	layout            *Layout
	keyBindingHandler KeyBindingHandler
}

func NewApp() (*App, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("creating gui: %w", err)
	}

	layout := NewLayout()
	keyBindingHandler := NewKeyBindingHandler(layout)

	g.Cursor = false
	g.Mouse = false
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout.Manager)

	if err := keyBindingHandler.SetupKeyBindings(g); err != nil {
		g.Close()
		return nil, fmt.Errorf("setting up key bindings: %w", err)
	}

	return &App{
		gui:               g,
		layout:            layout,
		keyBindingHandler: keyBindingHandler,
	}, nil
}

func (a *App) Run() error {
	defer a.gui.Close()

	if err := a.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Printf("main loop error: %v", err)
		return fmt.Errorf("running main loop: %w", err)
	}

	return nil
}
