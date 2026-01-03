package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type SchemaRegistryView struct {
	BaseView
	viewModel *viewmodel.SchemaRegistryViewModel
}

func NewSchemaRegistryView(vm *viewmodel.SchemaRegistryViewModel) *SchemaRegistryView {
	return &SchemaRegistryView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *SchemaRegistryView) Initialize(g *gocui.Gui) (bool, error) {
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

func (v *SchemaRegistryView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
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

func (v *SchemaRegistryView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *SchemaRegistryView) StartListening(g *gocui.Gui) {
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
