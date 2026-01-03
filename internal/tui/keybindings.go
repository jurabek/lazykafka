package tui

import (
	"errors"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type KeyBindingHandler interface {
	SetupKeyBindings(g *gocui.Gui) error
}

type keyBindingHandler struct {
	layout *Layout
}

func NewKeyBindingHandler(layout *Layout) KeyBindingHandler {
	return &keyBindingHandler{layout: layout}
}

func (h *keyBindingHandler) SetupKeyBindings(g *gocui.Gui) error {
	if err := h.setupGlobalBindings(g); err != nil {
		return err
	}

	for _, view := range h.layout.views {
		vm := view.GetViewModel()
		viewName := vm.GetName()
		bindings := vm.GetCommandBindings()

		if err := h.bindViewCommands(g, viewName, bindings); err != nil {
			return err
		}
	}

	return nil
}

func (h *keyBindingHandler) bindViewCommands(g *gocui.Gui, viewName string, bindings []*types.CommandBinding) error {
	for _, binding := range bindings {
		b := binding
		if err := g.SetKeybinding(viewName, b.Key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			err := b.Cmd.Execute()
			if err != nil && errors.Is(err, types.ErrNoSelection) {
				return nil
			}
			return err
		}); err != nil {

			return err
		}
	}
	return nil
}

func (h *keyBindingHandler) wrapHandler(handler func() error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return handler()
	}
}

func (h *keyBindingHandler) setupGlobalBindings(g *gocui.Gui) error {
	globalBindings := h.getGlobalBindings()
	for _, binding := range globalBindings {
		wrappedHandler := h.wrapHandler(binding.Handler)
		if err := g.SetKeybinding("", binding.Key, binding.Modifier, wrappedHandler); err != nil {
			return err
		}
	}
	return nil
}

func (h *keyBindingHandler) getGlobalBindings() []*types.Binding {
	return []*types.Binding{
		{
			ViewName:    "",
			Key:         gocui.KeyCtrlC,
			Modifier:    gocui.ModNone,
			Handler:     h.quit,
			Description: "quit",
		},
		{
			ViewName:    "",
			Key:         'q',
			Modifier:    gocui.ModNone,
			Handler:     h.quit,
			Description: "quit",
		},
		{
			ViewName:    "",
			Key:         gocui.KeyArrowRight,
			Modifier:    gocui.ModNone,
			Handler:     h.nextPanel,
			Description: "next panel",
		},
		{
			ViewName:    "",
			Key:         gocui.KeyArrowLeft,
			Modifier:    gocui.ModNone,
			Handler:     h.prevPanel,
			Description: "previous panel",
		},
		{
			ViewName:    "",
			Key:         'l',
			Modifier:    gocui.ModNone,
			Handler:     h.nextPanel,
			Description: "next panel",
		},
		{
			ViewName:    "",
			Key:         'h',
			Modifier:    gocui.ModNone,
			Handler:     h.prevPanel,
			Description: "previous panel",
		},
		{
			ViewName:    "",
			Key:         '1',
			Modifier:    gocui.ModNone,
			Handler:     h.jumpToPanel1,
			Description: "jump to panel 1",
		},
		{
			ViewName:    "",
			Key:         '2',
			Modifier:    gocui.ModNone,
			Handler:     h.jumpToPanel2,
			Description: "jump to panel 2",
		},
		{
			ViewName:    "",
			Key:         '3',
			Modifier:    gocui.ModNone,
			Handler:     h.jumpToPanel3,
			Description: "jump to panel 3",
		},
		{
			ViewName:    "",
			Key:         '4',
			Modifier:    gocui.ModNone,
			Handler:     h.jumpToPanel4,
			Description: "jump to panel 4",
		},
	}
}

func (h *keyBindingHandler) quit() error {
	return gocui.ErrQuit
}

func (h *keyBindingHandler) nextPanel() error {
	h.layout.NextPanel(h.layout.gui)
	return nil
}

func (h *keyBindingHandler) prevPanel() error {
	h.layout.PrevPanel(h.layout.gui)
	return nil
}

func (h *keyBindingHandler) jumpToPanel1() error {
	h.layout.JumpToPanel(h.layout.gui, 0)
	return nil
}

func (h *keyBindingHandler) jumpToPanel2() error {
	h.layout.JumpToPanel(h.layout.gui, 1)
	return nil
}

func (h *keyBindingHandler) jumpToPanel3() error {
	h.layout.JumpToPanel(h.layout.gui, 2)
	return nil
}

func (h *keyBindingHandler) jumpToPanel4() error {
	h.layout.JumpToPanel(h.layout.gui, 3)
	return nil
}
