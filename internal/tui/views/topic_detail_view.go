package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	tabsViewName = "topic_detail_tabs"
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

	tabsView, err := g.SetView(tabsViewName, x0, y0, x1, y0+2)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	view, err := g.SetView(v.viewModel.GetName(), x0, y0+2, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	created := err == gocui.ErrUnknownView
	if created {
		tabsView.Frame = false
		view.Title = v.viewModel.GetTitle()
		view.Wrap = false
	}

	return created, nil
}

func (v *TopicDetailView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	tabsView, err := g.View(tabsViewName)
	if err != nil {
		return err
	}

	v.renderTabs(tabsView)

	gocuiView.Clear()
	gocuiView.Title = v.viewModel.GetTitle()

	activeTab := v.viewModel.GetActiveTab()
	switch activeTab {
	case viewmodel.TabPartitions:
		maxX, _ := gocuiView.Size()
		content := v.viewModel.RenderPartitionsTable(maxX)
		fmt.Fprint(gocuiView, content)
	case viewmodel.TabConfiguration:
		fmt.Fprint(gocuiView, "  Configuration coming soon...")
	}

	return nil
}

func (v *TopicDetailView) renderTabs(tabsView *gocui.View) {
	tabsView.Clear()

	activeTab := v.viewModel.GetActiveTab()

	partitionsTab := "  Partitions  "
	configTab := "  Configuration  "

	if activeTab == viewmodel.TabPartitions {
		partitionsTab = " [Partitions] "
	} else {
		configTab = " [Configuration] "
	}

	fmt.Fprintf(tabsView, "%s%s", partitionsTab, configTab)
}

func (v *TopicDetailView) Destroy(g *gocui.Gui) error {
	if err := g.DeleteView(tabsViewName); err != nil && err != gocui.ErrUnknownView {
		return err
	}
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
