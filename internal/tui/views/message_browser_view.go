package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type MessageBrowserView struct {
	BaseView
	viewModel *viewmodel.MessageBrowserViewModel
}

func NewMessageBrowserView(vm *viewmodel.MessageBrowserViewModel) *MessageBrowserView {
	return &MessageBrowserView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *MessageBrowserView) Initialize(g *gocui.Gui) (bool, error) {
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

func (v *MessageBrowserView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	gocuiView.Highlight = v.IsActive()

	items := v.viewModel.GetDisplayItems()
	selectedIdx := v.viewModel.GetSelectedIndex()

	for i, item := range items {
		if i == selectedIdx && v.IsActive() {
			gocuiView.SetCursor(0, i)
			fmt.Fprintf(gocuiView, "> %s\n", item)
		} else {
			fmt.Fprintf(gocuiView, "  %s\n", item)
		}
	}

	if len(items) == 0 {
		fmt.Fprintln(gocuiView, "  No messages")
	}

	return nil
}

func (v *MessageBrowserView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *MessageBrowserView) SetupCallbacks(g *gocui.Gui) {
	renderFn := func() {
		g.Update(func(gui *gocui.Gui) error {
			view, err := g.View(v.viewModel.GetName())
			if err != nil {
				return nil
			}
			gocuiView, _ := g.View(v.viewModel.GetName())
			if gocuiView != nil {
				gocuiView.Title = v.viewModel.GetTitle()
			}
			return v.Render(g, view)
		})
	}

	for _, binding := range v.viewModel.GetCommandBindings() {
		binding.Cmd.SetOnExecuted(renderFn)
	}

	v.viewModel.SetOnChange(func(event types.ChangeEvent) {
		renderFn()
	})
}
