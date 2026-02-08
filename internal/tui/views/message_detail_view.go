package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type MessageDetailView struct {
	BaseView
	viewModel *viewmodel.MessageDetailViewModel
}

func NewMessageDetailView(vm *viewmodel.MessageDetailViewModel) *MessageDetailView {
	return &MessageDetailView{
		BaseView:  BaseView{viewModel: vm},
		viewModel: vm,
	}
}

func (v *MessageDetailView) Initialize(g *gocui.Gui) (bool, error) {
	x0, y0, x1, y1 := v.GetBounds()

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return false, err
	}

	created := err == gocui.ErrUnknownView
	if created {
		view.Title = v.viewModel.GetTitle()
		view.Wrap = false
		view.Editable = false
		view.Autoscroll = false
	}

	return created, nil
}

func (v *MessageDetailView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	gocuiView.Title = v.viewModel.GetTitle()

	lines := v.viewModel.GetDisplayText()
	scrollPosition := v.viewModel.GetScrollPosition()

	_, viewHeight := gocuiView.Size()

	startIdx := scrollPosition
	maxStartIdx := len(lines) - viewHeight
	if maxStartIdx < 0 {
		maxStartIdx = 0
	}
	if startIdx > maxStartIdx {
		startIdx = maxStartIdx
	}

	endIdx := startIdx + viewHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	for i := startIdx; i < endIdx; i++ {
		if i < len(lines) {
			fmt.Fprintln(gocuiView, lines[i])
		}
	}

	if len(lines) == 0 {
		fmt.Fprint(gocuiView, "No message selected")
	}

	_, originY := gocuiView.Origin()
	if originY != startIdx {
		gocuiView.SetOrigin(0, startIdx)
	}

	return nil
}

func (v *MessageDetailView) Destroy(g *gocui.Gui) error {
	return g.DeleteView(v.viewModel.GetName())
}

func (v *MessageDetailView) SetupCallbacks(g *gocui.Gui) {
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
