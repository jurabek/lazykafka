package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type ConsumerGroupsView struct {
	BaseView
	viewModel *viewmodel.ConsumerGroupsViewModel
}

func NewConsumerGroupsView(vm *viewmodel.ConsumerGroupsViewModel) *ConsumerGroupsView {
	return &ConsumerGroupsView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *ConsumerGroupsView) Initialize(g *gocui.Gui) error {
	x0, y0, x1, y1 := v.GetBounds()

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = v.viewModel.GetTitle()
	view.Highlight = true
	view.SelBgColor = gocui.ColorGreen
	view.SelFgColor = gocui.ColorBlack

	return v.Render(g, view)
}

func (v *ConsumerGroupsView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()

	items := v.viewModel.GetDisplayItems()
	selectedIdx := v.viewModel.GetSelectedIndex()

	for i, item := range items {
		if i == selectedIdx {
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

func (v *ConsumerGroupsView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}
