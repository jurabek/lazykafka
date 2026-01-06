package views

import (
	"fmt"
	"log/slog"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type BrokersView struct {
	BaseView
	viewModel    *viewmodel.BrokersViewModel
	scrollOffset int
}

func NewBrokersView(vm *viewmodel.BrokersViewModel) *BrokersView {
	return &BrokersView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *BrokersView) Initialize(g *gocui.Gui) (bool, error) {
	x0, y0, x1, y1 := v.GetBounds()

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	created := err == gocui.ErrUnknownView
	if created {
		view.Title = v.viewModel.GetTitle()
		view.Highlight = true
		view.SelBgColor = gocui.ColorBlue
		view.SelFgColor = gocui.ColorBlack
	}

	return created, nil
}

func (v *BrokersView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	slog.Info("rendering brokers view")
	gocuiView.Clear()
	gocuiView.Highlight = v.IsActive()

	items := v.viewModel.GetDisplayItems()
	selectedIdx := v.viewModel.GetSelectedIndex()

	// Calculate view height
	_, viewHeight := gocuiView.Size()

	// Ensure scroll offset is valid
	totalItems := len(items)
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
	if v.scrollOffset > totalItems-viewHeight && totalItems > viewHeight {
		v.scrollOffset = totalItems - viewHeight
	}
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}

	// Adjust scroll offset if selected item is out of view
	if selectedIdx >= 0 {
		if selectedIdx < v.scrollOffset {
			v.scrollOffset = selectedIdx
		} else if selectedIdx >= v.scrollOffset+viewHeight {
			v.scrollOffset = selectedIdx - viewHeight + 1
		}
	}

	// Render visible items
	visibleItems := items
	if totalItems > viewHeight {
		endIdx := v.scrollOffset + viewHeight
		if endIdx > totalItems {
			endIdx = totalItems
		}
		visibleItems = items[v.scrollOffset:endIdx]
	}

	for i, item := range visibleItems {
		actualIndex := i + v.scrollOffset
		if actualIndex == selectedIdx {
			gocuiView.SetCursor(0, i)
			fmt.Fprintf(gocuiView, "> %s\n", item)
		} else {
			fmt.Fprintf(gocuiView, "  %s\n", item)
		}
	}

	if len(items) == 0 {
		fmt.Fprintln(gocuiView, "  (empty)")
	}

	return nil
}

func (v *BrokersView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *BrokersView) ScrollUp() error {
	if v.scrollOffset > 0 {
		v.scrollOffset--
		return nil
	}
	return types.ErrNoSelection
}

func (v *BrokersView) ScrollDown() error {
	_, y0, _, y1 := v.GetBounds()
	viewHeight := y1 - y0 + 1 // Calculate actual height

	totalItems := v.viewModel.GetItemCount()
	if v.scrollOffset < totalItems-viewHeight {
		v.scrollOffset++
		return nil
	}
	return types.ErrNoSelection
}

func (v *BrokersView) SetupCallbacks(g *gocui.Gui) {
	scrollUpCmd := types.NewCommand(v.ScrollUp)
	scrollDownCmd := types.NewCommand(v.ScrollDown)

	renderFn := func() {
		g.Update(func(gui *gocui.Gui) error {
			view, err := g.View(v.viewModel.GetName())
			if err != nil {
				return nil
			}
			return v.Render(g, view)
		})
	}

	for _, binding := range v.viewModel.GetCommandBindings() {
		binding.Cmd.SetOnExecuted(renderFn)
	}

	scrollUpCmd.SetOnExecuted(renderFn)
	scrollDownCmd.SetOnExecuted(renderFn)

	v.viewModel.SetOnChange(func(event types.ChangeEvent) {
		renderFn()
	})

	// Add scroll keybindings (Ctrl+U and Ctrl+D for scrolling)
	if err := g.SetKeybinding(v.viewModel.GetName(), gocui.KeyCtrlU, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return scrollUpCmd.Execute()
	}); err != nil {
		slog.Error("failed to set Ctrl+U keybinding", "error", err)
	}
	if err := g.SetKeybinding(v.viewModel.GetName(), gocui.KeyCtrlD, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return scrollDownCmd.Execute()
	}); err != nil {
		slog.Error("failed to set Ctrl+D keybinding", "error", err)
	}
}
