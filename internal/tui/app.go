package tui

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/jroimartin/gocui"
)

type App struct {
	gui               *gocui.Gui
	layout            *Layout
	keyBindingHandler KeyBindingHandler
	ctx               context.Context
}

func NewApp(ctx context.Context) (*App, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("creating gui: %w", err)
	}

	layout := NewLayout(ctx, g)
	keyBindingHandler := NewKeyBindingHandler(layout)

	g.Cursor = false
	g.Mouse = false
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.InputEsc = true

	g.SetManagerFunc(layout.Manager)

	if err := keyBindingHandler.SetupKeyBindings(g); err != nil {
		g.Close()
		return nil, fmt.Errorf("setting up key bindings: %w", err)
	}

	return &App{
		gui:               g,
		layout:            layout,
		keyBindingHandler: keyBindingHandler,
		ctx:               ctx,
	}, nil
}

func (a *App) Run() error {
	defer a.gui.Close()
	defer a.ctx.Done()

	defer slog.Info("App closed")

	if err := a.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Printf("main loop error: %v", err)
		return fmt.Errorf("running main loop: %w", err)
	}

	return nil
}
