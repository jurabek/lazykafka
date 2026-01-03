package tui

import (
	"context"
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
	"github.com/jurabek/lazykafka/internal/tui/views"
)

type Layout struct {
	views           []views.View
	activeViewIndex int
	gui             *gocui.Gui
	mainVM          *viewmodel.MainViewModel
}

func NewLayout(ctx context.Context, g *gocui.Gui) *Layout {
	mainVM := viewmodel.NewMainViewModel(ctx)

	brokersView := views.NewBrokersView(mainVM.BrokersVM())
	topicsView := views.NewTopicsView(mainVM.TopicsVM())
	cgView := views.NewConsumerGroupsView(mainVM.ConsumerGroupsVM())
	srView := views.NewSchemaRegistryView(mainVM.SchemaRegistryVM())

	viewList := []views.View{brokersView, topicsView, cgView, srView}

	for _, v := range viewList {
		v.StartListening(g)
	}

	return &Layout{
		views:           viewList,
		activeViewIndex: 0,
		gui:             g,
		mainVM:          mainVM,
	}
}

func (l *Layout) Manager(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if maxX < minTermWidth || maxY < minTermHeight {
		return fmt.Errorf("terminal too small: need at least %dx%d", minTermWidth, minTermHeight)
	}

	panelHeight := (maxY - 3) / len(l.views)
	helpHeight := 2

	for i, view := range l.views {
		y0 := i * panelHeight
		y1 := y0 + panelHeight - 1
		if i == len(l.views)-1 {
			y1 = maxY - helpHeight - 2
		}

		view.SetBounds(0, y0, maxX-1, y1)

		if err := view.Initialize(g); err != nil {
			return fmt.Errorf("initializing view %s: %w", view.GetViewModel().GetName(), err)
		}

		if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
			if err := view.Render(g, gocuiView); err != nil {
				return fmt.Errorf("rendering view: %w", err)
			}
		}
	}

	if err := l.createHelpView(g, maxX, maxY, helpHeight); err != nil {
		return fmt.Errorf("creating help view: %w", err)
	}

	activeView := l.views[l.activeViewIndex]
	if _, err := g.SetCurrentView(activeView.GetViewModel().GetName()); err != nil {
		return fmt.Errorf("setting current view: %w", err)
	}

	return nil
}

func (l *Layout) createHelpView(g *gocui.Gui, maxX, maxY, helpHeight int) error {
	v, err := g.SetView(panelHelp, 0, maxY-helpHeight-1, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Frame = false
	v.Clear()
	fmt.Fprintln(v, " ←/→: switch panel | ↑/k: up | ↓/j: down | 1-4: jump panel | n: new | q: quit")
	return nil
}

func (l *Layout) NextPanel(g *gocui.Gui) {
	l.activeViewIndex = (l.activeViewIndex + 1) % len(l.views)
	l.refreshAllViews(g)
}

func (l *Layout) PrevPanel(g *gocui.Gui) {
	l.activeViewIndex--
	if l.activeViewIndex < 0 {
		l.activeViewIndex = len(l.views) - 1
	}
	l.refreshAllViews(g)
}

func (l *Layout) JumpToPanel(g *gocui.Gui, index int) {
	if index >= 0 && index < len(l.views) {
		l.activeViewIndex = index
		l.refreshAllViews(g)
	}
}

func (l *Layout) refreshAllViews(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		activeView := l.views[l.activeViewIndex]
		_, _ = g.SetCurrentView(activeView.GetViewModel().GetName())
		return nil
	})
}

func (l *Layout) MainViewModel() *viewmodel.MainViewModel {
	return l.mainVM
}
