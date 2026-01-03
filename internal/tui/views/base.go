package views

import (
	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type View interface {
	Initialize(g *gocui.Gui) error
	Render(g *gocui.Gui, v *gocui.View) error
	Destroy(g *gocui.Gui) error
	GetViewModel() viewmodel.BaseViewModel
	GetBounds() (x0, y0, x1, y1 int)
	SetBounds(x0, y0, x1, y1 int)
	StartListening(g *gocui.Gui)
}

type BaseView struct {
	viewModel      viewmodel.BaseViewModel
	x0, y0, x1, y1 int
}

func (v *BaseView) GetViewModel() viewmodel.BaseViewModel {
	return v.viewModel
}

func (v *BaseView) GetBounds() (x0, y0, x1, y1 int) {
	return v.x0, v.y0, v.x1, v.y1
}

func (v *BaseView) SetBounds(x0, y0, x1, y1 int) {
	v.x0 = x0
	v.y0 = y0
	v.x1 = x1
	v.y1 = y1
}
