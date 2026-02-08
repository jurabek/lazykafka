package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type ConfirmView struct {
	viewModel *viewmodel.ConfirmViewModel
	onClose   func()
}

func NewConfirmView(vm *viewmodel.ConfirmViewModel, onClose func()) *ConfirmView {
	return &ConfirmView{
		viewModel: vm,
		onClose:   onClose,
	}
}

func (v *ConfirmView) Initialize(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	width := 50
	height := 5
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	view, err := g.SetView(v.viewModel.GetName(), x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = v.viewModel.GetTitle()
	view.Frame = true
	view.Editable = false

	if err := g.SetKeybinding(v.viewModel.GetName(), 'y', gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.viewModel.Confirm()
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(v.viewModel.GetName(), 'n', gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.viewModel.Cancel()
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(v.viewModel.GetName(), gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.viewModel.Cancel()
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (v *ConfirmView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	message := v.viewModel.GetMessage()
	fmt.Fprintf(gocuiView, "\n  %s (y/n)", message)
	return nil
}

func (v *ConfirmView) Destroy(g *gocui.Gui) error {
	_ = g.DeleteView(v.viewModel.GetName())
	g.DeleteKeybindings(v.viewModel.GetName())
	return nil
}

func (v *ConfirmView) GetViewModel() *viewmodel.ConfirmViewModel {
	return v.viewModel
}

func (v *ConfirmView) SetupCallbacks(g *gocui.Gui) {}
func (v *ConfirmView) SetActive(bool)              {}
