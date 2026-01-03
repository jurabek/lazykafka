package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type TopicsView struct {
	BaseView
	viewModel *viewmodel.TopicsViewModel
}

func NewTopicsView(vm *viewmodel.TopicsViewModel) *TopicsView {
	return &TopicsView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *TopicsView) Initialize(g *gocui.Gui) (bool, error) {
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

func (v *TopicsView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
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
		fmt.Fprintln(gocuiView, "  (empty)")
	}

	return nil
}

func (v *TopicsView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *TopicsView) StartListening(g *gocui.Gui) {
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
