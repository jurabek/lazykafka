package views

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	wizardPopup = "wizard_popup"
	wizardInput = "wizard_input"

	wizardWidth  = 100
	wizardHeight = 2
)

type wizardEditor struct {
	onEsc       func()
	onEnter     func()
	onArrowUp   func()
	onArrowDown func()
	view        *AddBrokerView
}

func (e *wizardEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyEsc:
		if e.onEsc != nil {
			e.onEsc()
		}
		return
	case gocui.KeyEnter:
		if e.onEnter != nil {
			e.onEnter()
		}
		return
	case gocui.KeyArrowUp:
		if e.onArrowUp != nil {
			e.onArrowUp()
		}
		return
	case gocui.KeyArrowDown:
		if e.onArrowDown != nil {
			e.onArrowDown()
		}
		return
	}

	// Prevent text input during list selection steps
	if e.view != nil {
		step := e.view.viewModel.GetCurrentStep()
		if step == viewmodel.StepAuthType || step == viewmodel.StepSASLMechanism {
			return
		}
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

type AddBrokerView struct {
	viewModel *viewmodel.AddBrokerViewModel
	gui       *gocui.Gui
	onClose   func()
}

func NewAddBrokerView(vm *viewmodel.AddBrokerViewModel, onClose func()) *AddBrokerView {
	return &AddBrokerView{
		viewModel: vm,
		onClose:   onClose,
	}
}

func (v *AddBrokerView) GetViewModel() *viewmodel.AddBrokerViewModel {
	return v.viewModel
}

func (v *AddBrokerView) Initialize(g *gocui.Gui) error {
	v.gui = g
	return v.render()
}

func (v *AddBrokerView) render() error {
	maxX, maxY := v.gui.Size()

	// Calculate dynamic height
	step := v.viewModel.GetCurrentStep()
	var wizardHeight int
	switch step {
	case viewmodel.StepAuthType:
		wizardHeight = 4 // 2 options + title + padding
	case viewmodel.StepSASLMechanism:
		wizardHeight = 6 // 4 options + title + padding
	default:
		wizardHeight = 2 // standard input height
	}

	x0 := (maxX - wizardWidth) / 2
	y0 := (maxY - wizardHeight) / 2
	x1 := x0 + wizardWidth
	y1 := y0 + wizardHeight

	inputView, err := v.gui.SetView(wizardInput, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	inputView.Title = " " + v.viewModel.GetStepTitle() + " "
	inputView.Editable = true
	inputView.Editor = &wizardEditor{
		onEsc:       v.handleEsc,
		onEnter:     v.handleEnter,
		onArrowUp:   v.handleArrowUp,
		onArrowDown: v.handleArrowDown,
		view:        v,
	}

	// Handle different step rendering
	switch step {
	case viewmodel.StepAuthType:
		v.renderAuthTypeList(inputView)
		v.gui.Cursor = false
	case viewmodel.StepSASLMechanism:
		v.renderSASLMechanismList(inputView)
		v.gui.Cursor = false
	default:
		inputView.SetCursor(0, 0)
		v.gui.Cursor = true
		if step == viewmodel.StepPassword {
			inputView.Mask = '*'
		} else {
			inputView.Mask = 0
		}
	}

	_, _ = v.gui.SetViewOnTop(wizardInput)

	if _, err := v.gui.SetCurrentView(wizardInput); err != nil {
		slog.Error("failed to set current view", "view", wizardInput, "error", err)
	} else {
		slog.Info("set current view", "view", wizardInput)
	}

	v.gui.Cursor = true

	// Verify current view
	if cv := v.gui.CurrentView(); cv != nil {
		slog.Info("current view after set", "name", cv.Name(), "editable", cv.Editable)
	}

	return nil
}

func (v *AddBrokerView) handleEsc() {
	v.viewModel.Cancel()
}

func (v *AddBrokerView) handleEnter() {
	step := v.viewModel.GetCurrentStep()

	if step == viewmodel.StepAuthType || step == viewmodel.StepSASLMechanism {
		if v.viewModel.NextStep() {
			_ = v.viewModel.Submit()
		} else {
			v.clearAndRender()
		}
		return
	}

	v.saveCurrentValue()

	if v.viewModel.NextStep() {
		_ = v.viewModel.Submit()
	} else {
		v.clearAndRender()
	}
}

func (v *AddBrokerView) handleArrowUp() {
	step := v.viewModel.GetCurrentStep()
	switch step {
	case viewmodel.StepAuthType:
		v.viewModel.MoveAuthTypeUp()
		v.updateAuthDisplay()
	case viewmodel.StepSASLMechanism:
		v.viewModel.MoveSASLMechanismUp()
		v.updateSASLMechanismDisplay()
	}
}

func (v *AddBrokerView) handleArrowDown() {
	step := v.viewModel.GetCurrentStep()
	switch step {
	case viewmodel.StepAuthType:
		v.viewModel.MoveAuthTypeDown()
		v.updateAuthDisplay()
	case viewmodel.StepSASLMechanism:
		v.viewModel.MoveSASLMechanismDown()
		v.updateSASLMechanismDisplay()
	}
}

func (v *AddBrokerView) renderAuthTypeList(inputView *gocui.View) {
	inputView.Clear()
	options := v.viewModel.GetAuthTypeOptions()
	selectedIdx := v.viewModel.GetSelectedAuthTypeIndex()

	for i, option := range options {
		prefix := "  "
		if i == selectedIdx {
			prefix = "> "
		}
		fmt.Fprintf(inputView, "%s%s\n", prefix, option)
	}
}

func (v *AddBrokerView) renderSASLMechanismList(inputView *gocui.View) {
	inputView.Clear()
	options := v.viewModel.GetSASLMechanismOptions()
	selectedIdx := v.viewModel.GetSelectedSASLMechanismIndex()

	for i, option := range options {
		prefix := "  "
		if i == selectedIdx {
			prefix = "> "
		}
		fmt.Fprintf(inputView, "%s%s\n", prefix, option)
	}
}

func (v *AddBrokerView) updateSASLMechanismDisplay() {
	inputView, err := v.gui.View(wizardInput)
	if err != nil {
		return
	}
	v.renderSASLMechanismList(inputView)
}

func (v *AddBrokerView) saveCurrentValue() {
	inputView, err := v.gui.View(wizardInput)
	if err != nil {
		return
	}
	value := strings.TrimSpace(inputView.Buffer())

	switch v.viewModel.GetCurrentStep() {
	case viewmodel.StepName:
		v.viewModel.SetName(value)
	case viewmodel.StepBootstrapServers:
		v.viewModel.SetBootstrapServers(value)
	case viewmodel.StepUsername:
		v.viewModel.SetUsername(value)
	case viewmodel.StepPassword:
		v.viewModel.SetPassword(value)
	}
}

func (v *AddBrokerView) updateAuthDisplay() {
	inputView, err := v.gui.View(wizardInput)
	if err != nil {
		return
	}
	v.renderAuthTypeList(inputView)
}

func (v *AddBrokerView) clearAndRender() {
	inputView, err := v.gui.View(wizardInput)
	if err != nil {
		return
	}
	inputView.Clear()
	inputView.SetCursor(0, 0)
	_ = v.render()
}

func (v *AddBrokerView) Destroy(g *gocui.Gui) error {
	g.Cursor = false
	_ = g.DeleteView(wizardInput)
	_ = g.DeleteView(wizardPopup)
	return nil
}
