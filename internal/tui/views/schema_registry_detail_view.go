package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type SchemaRegistryDetailView struct {
	BaseView
	viewModel *viewmodel.SchemaRegistryDetailViewModel
}

func NewSchemaRegistryDetailView(vm *viewmodel.SchemaRegistryDetailViewModel) *SchemaRegistryDetailView {
	return &SchemaRegistryDetailView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *SchemaRegistryDetailView) Initialize(g *gocui.Gui) (bool, error) {
	x0, y0, x1, y1 := v.GetBounds()

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	created := err == gocui.ErrUnknownView
	if created {
		view.Title = v.viewModel.GetTitle()
		view.Wrap = true
	}

	return created, nil
}

func (v *SchemaRegistryDetailView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	gocuiView.Title = v.viewModel.GetTitle()

	schemaType := v.viewModel.GetSchemaType()
	if schemaType != "" {
		fmt.Fprintf(gocuiView, "Type: %s\n\n", schemaType)
	}

	content := v.viewModel.GetSchemaContent()
	fmt.Fprint(gocuiView, content)

	return nil
}

func (v *SchemaRegistryDetailView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *SchemaRegistryDetailView) SetupCallbacks(g *gocui.Gui) {
	v.viewModel.SetOnChange(func(event types.ChangeEvent) {
		g.Update(func(gui *gocui.Gui) error {
			view, err := g.View(v.viewModel.GetName())
			if err != nil {
				return nil
			}
			return v.Render(g, view)
		})
	})
}
