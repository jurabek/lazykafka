package tui

import "github.com/jroimartin/gocui"

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
	// Global quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, h.quit); err != nil {
		return err
	}

	// Arrow keys work globally
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, h.moveUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, h.moveDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, h.nextPanel); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, h.prevPanel); err != nil {
		return err
	}

	// Panel-specific bindings
	panels := []string{panelBrokers, panelTopics, panelConsumerGroups, panelSchemaRegistry}
	for _, panel := range panels {
		if err := g.SetKeybinding(panel, 'q', gocui.ModNone, h.quit); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, 'k', gocui.ModNone, h.moveUp); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, 'j', gocui.ModNone, h.moveDown); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, 'l', gocui.ModNone, h.nextPanel); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, 'h', gocui.ModNone, h.prevPanel); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, '1', gocui.ModNone, h.jumpToPanel1); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, '2', gocui.ModNone, h.jumpToPanel2); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, '3', gocui.ModNone, h.jumpToPanel3); err != nil {
			return err
		}
		if err := g.SetKeybinding(panel, '4', gocui.ModNone, h.jumpToPanel4); err != nil {
			return err
		}
	}

	return nil
}

func (h *keyBindingHandler) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (h *keyBindingHandler) moveUp(g *gocui.Gui, _ *gocui.View) error {
	h.layout.MoveUp(g)
	return nil
}

func (h *keyBindingHandler) moveDown(g *gocui.Gui, _ *gocui.View) error {
	h.layout.MoveDown(g)
	return nil
}

func (h *keyBindingHandler) nextPanel(g *gocui.Gui, _ *gocui.View) error {
	h.layout.NextPanel(g)
	return nil
}

func (h *keyBindingHandler) prevPanel(g *gocui.Gui, _ *gocui.View) error {
	h.layout.PrevPanel(g)
	return nil
}

func (h *keyBindingHandler) jumpToPanel1(g *gocui.Gui, _ *gocui.View) error {
	h.layout.JumpToPanel(g, 0)
	return nil
}

func (h *keyBindingHandler) jumpToPanel2(g *gocui.Gui, _ *gocui.View) error {
	h.layout.JumpToPanel(g, 1)
	return nil
}

func (h *keyBindingHandler) jumpToPanel3(g *gocui.Gui, _ *gocui.View) error {
	h.layout.JumpToPanel(g, 2)
	return nil
}

func (h *keyBindingHandler) jumpToPanel4(g *gocui.Gui, _ *gocui.View) error {
	h.layout.JumpToPanel(g, 3)
	return nil
}
