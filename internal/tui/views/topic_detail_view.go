package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)


type TopicDetailView struct {
	BaseView
	viewModel *viewmodel.TopicDetailViewModel
}

func NewTopicDetailView(vm *viewmodel.TopicDetailViewModel) *TopicDetailView {
	return &TopicDetailView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *TopicDetailView) Initialize(g *gocui.Gui) (bool, error) {
	x0, y0, x1, y1 := v.GetBounds()

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	created := err == gocui.ErrUnknownView
	if created {
		view.Title = v.viewModel.GetTitle()
		view.Wrap = false
	}

	return created, nil
}

func (v *TopicDetailView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	gocuiView.Title = v.viewModel.GetTitle()

	maxX, _ := gocuiView.Size()
	content := v.viewModel.RenderPartitionsTable(maxX)
	fmt.Fprint(gocuiView, content)

	return nil
}

func (v *TopicDetailView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *TopicDetailView) StartListening(g *gocui.Gui) {
	go func() {
		for range v.viewModel.NotifyChannel() {
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
