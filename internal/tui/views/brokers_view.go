package views

import (
	"fmt"
	"log/slog"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type BrokersView struct {
	BaseView
	viewModel *viewmodel.BrokersViewModel
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

	items := v.viewModel.GetDisplayItems()
	selectedIdx := v.viewModel.GetSelectedIndex()

	for i, item := range items {
		if i == selectedIdx {
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

func (v *BrokersView) StartListening(g *gocui.Gui) {
	for _, binding := range v.viewModel.GetCommandBindings() {
		cmd := binding.Cmd
		go func() {
			for range cmd.NotifyChannel() {
				g.Update(func(gui *gocui.Gui) error {
					view, err := g.View(v.viewModel.GetName())
					if err != nil {
						return err
					}
					return v.Render(g, view)
				})
			}
		}()
	}
}
