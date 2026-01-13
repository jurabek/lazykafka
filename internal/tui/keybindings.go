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

	for _, view := range h.layout.sidebarViews {
		vm := view.GetViewModel()
		viewName := vm.GetName()
		bindings := vm.GetCommandBindings()

		if err := h.bindViewCommands(g, viewName, bindings); err != nil {
			return err
		}

		// Bind h/l/q for sidebar views only (not consumed in popup)
		if err := g.SetKeybinding(viewName, 'h', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			h.layout.PrevPanel(h.layout.gui)
			return nil
		}); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, 'l', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			h.layout.NextPanel(h.layout.gui)
			return nil
		}); err != nil {
			return err
		}
		if err := g.SetKeybinding(viewName, 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
			return err
		}

		// Bind panel navigation keys per-view so they don't consume keys in popup
		if err := h.bindPanelNavigationKeys(g, viewName); err != nil {
			return err
		}
	}

	return nil
}

func (h *keyBindingHandler) bindViewCommands(g *gocui.Gui, viewName string, bindings []*types.CommandBinding) error {
	for _, binding := range bindings {
		b := binding
		if err := g.SetKeybinding(viewName, b.Key, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			if h.layout.IsPopupActive() {
				return nil
			}
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

func (h *keyBindingHandler) wrapHandler(handler func() error, blockOnPopup bool) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if blockOnPopup && h.layout.IsPopupActive() {
			return nil
		}
		return handler()
	}
}

func (h *keyBindingHandler) setupGlobalBindings(g *gocui.Gui) error {
	globalBindings := h.getGlobalBindings()
	for _, binding := range globalBindings {
		wrappedHandler := h.wrapHandler(binding.Handler, binding.BlockOnPopup)
		if err := g.SetKeybinding(binding.ViewName, binding.Key, binding.Modifier, wrappedHandler); err != nil {
			return err
		}
	}
	return nil
}

func (h *keyBindingHandler) getGlobalBindings() []*types.Binding {
	// Only truly global bindings that should work everywhere
	// Panel navigation keys (1,2,3,4, arrows) are bound per-view in bindPanelNavigationKeys
	return []*types.Binding{
		{
			ViewName:     "",
			Key:          gocui.KeyCtrlC,
			Modifier:     gocui.ModNone,
			Handler:      h.quit,
			Description:  "quit",
			BlockOnPopup: false,
		},
		{
			ViewName:     panelBrokers,
			Key:          'n',
			Modifier:     gocui.ModNone,
			Handler:      h.showAddBrokerPopup,
			Description:  "new broker",
			BlockOnPopup: true,
		},
		{
			ViewName:     panelTopics,
			Key:          'n',
			Modifier:     gocui.ModNone,
			Handler:      h.showAddTopicPopup,
			Description:  "new topic",
			BlockOnPopup: true,
		},
	}
}

// bindPanelNavigationKeys binds panel navigation keys to a specific view
// This ensures they don't consume key events in the popup input view
func (h *keyBindingHandler) bindPanelNavigationKeys(g *gocui.Gui, viewName string) error {
	// Arrow keys for panel navigation
	if err := g.SetKeybinding(viewName, gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.NextPanel(h.layout.gui)
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.PrevPanel(h.layout.gui)
		return nil
	}); err != nil {
		return err
	}

	// Number keys to jump to panels
	if err := g.SetKeybinding(viewName, '1', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.JumpToPanel(h.layout.gui, 0)
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, '2', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.JumpToPanel(h.layout.gui, 1)
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, '3', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.JumpToPanel(h.layout.gui, 2)
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, '4', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		h.layout.JumpToPanel(h.layout.gui, 3)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (h *keyBindingHandler) quit() error {
	return gocui.ErrQuit
}

func (h *keyBindingHandler) showAddBrokerPopup() error {
	if h.layout.IsPopupActive() {
		return nil
	}
	return h.layout.ShowAddBrokerPopup()
}

func (h *keyBindingHandler) showAddTopicPopup() error {
	if h.layout.IsPopupActive() {
		return nil
	}
	return h.layout.ShowAddTopicPopup()
}
