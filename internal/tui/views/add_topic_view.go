package views

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const topicWizardInput = "topic_wizard_input"

type topicWizardEditor struct {
	onEsc       func()
	onEnter     func()
	onArrowUp   func()
	onArrowDown func()
	view        *AddTopicView
}

func (e *topicWizardEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
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

	if e.view != nil {
		step := e.view.viewModel.GetCurrentStep()
		if step == viewmodel.StepCleanupPolicy {
			return
		}
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

type AddTopicView struct {
	viewModel *viewmodel.AddTopicViewModel
	gui       *gocui.Gui
}

func NewAddTopicView(vm *viewmodel.AddTopicViewModel, _ func()) *AddTopicView {
	return &AddTopicView{
		viewModel: vm,
	}
}

func (v *AddTopicView) GetViewModel() *viewmodel.AddTopicViewModel {
	return v.viewModel
}

func (v *AddTopicView) Initialize(g *gocui.Gui) error {
	v.gui = g
	return v.render()
}

func (v *AddTopicView) render() error {
	maxX, maxY := v.gui.Size()

	step := v.viewModel.GetCurrentStep()
	var height int
	switch step {
	case viewmodel.StepCleanupPolicy:
		height = 5
	default:
		height = 2
	}

	x0 := (maxX - wizardWidth) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + wizardWidth
	y1 := y0 + height

	inputView, err := v.gui.SetView(topicWizardInput, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	inputView.Title = " " + v.viewModel.GetStepTitle() + " "
	inputView.Editable = true
	inputView.Editor = &topicWizardEditor{
		onEsc:       v.handleEsc,
		onEnter:     v.handleEnter,
		onArrowUp:   v.handleArrowUp,
		onArrowDown: v.handleArrowDown,
		view:        v,
	}

	switch step {
	case viewmodel.StepCleanupPolicy:
		v.renderCleanupPolicyList(inputView)
		v.gui.Cursor = false
	case viewmodel.StepMinISR:
		inputView.SetCursor(0, 0)
		v.gui.Cursor = true
		inputView.Mask = 0
		if inputView.Buffer() == "" {
			fmt.Fprint(inputView, v.viewModel.GetMinISR())
			inputView.SetCursor(len(v.viewModel.GetMinISR()), 0)
		}
	default:
		inputView.SetCursor(0, 0)
		v.gui.Cursor = true
		inputView.Mask = 0
	}

	_, _ = v.gui.SetViewOnTop(topicWizardInput)

	if _, err := v.gui.SetCurrentView(topicWizardInput); err != nil {
		slog.Error("failed to set current view", "view", topicWizardInput, "error", err)
	}

	v.gui.Cursor = true

	return nil
}

func (v *AddTopicView) handleEsc() {
	v.viewModel.Cancel()
}

func (v *AddTopicView) handleEnter() {
	if v.viewModel.GetCurrentStep() != viewmodel.StepCleanupPolicy {
		v.saveCurrentValue()
	}

	if v.viewModel.NextStep() {
		err := v.viewModel.Submit()
		if err != nil {
			slog.Error("failed to submit topic", "error", err)
		}
	} else {
		v.clearAndRender()
	}
}

func (v *AddTopicView) handleArrowUp() {
	step := v.viewModel.GetCurrentStep()
	if step == viewmodel.StepCleanupPolicy {
		v.viewModel.MoveCleanupPolicyUp()
		v.updateCleanupPolicyDisplay()
	}
}

func (v *AddTopicView) handleArrowDown() {
	step := v.viewModel.GetCurrentStep()
	if step == viewmodel.StepCleanupPolicy {
		v.viewModel.MoveCleanupPolicyDown()
		v.updateCleanupPolicyDisplay()
	}
}

func (v *AddTopicView) renderCleanupPolicyList(inputView *gocui.View) {
	inputView.Clear()
	options := v.viewModel.GetCleanupPolicyOptions()
	selectedIdx := v.viewModel.GetSelectedCleanupPolicyIndex()

	for i, option := range options {
		prefix := "  "
		if i == selectedIdx {
			prefix = "> "
		}
		fmt.Fprintf(inputView, "%s%s\n", prefix, option)
	}
}

func (v *AddTopicView) updateCleanupPolicyDisplay() {
	inputView, err := v.gui.View(topicWizardInput)
	if err != nil {
		return
	}
	v.renderCleanupPolicyList(inputView)
}

func (v *AddTopicView) saveCurrentValue() {
	inputView, err := v.gui.View(topicWizardInput)
	if err != nil {
		return
	}
	value := strings.TrimSpace(inputView.Buffer())

	switch v.viewModel.GetCurrentStep() {
	case viewmodel.StepTopicName:
		v.viewModel.SetName(value)
	case viewmodel.StepPartitions:
		v.viewModel.SetPartitions(value)
	case viewmodel.StepReplicationFactor:
		v.viewModel.SetReplicationFactor(value)
	case viewmodel.StepMinISR:
		v.viewModel.SetMinISR(value)
	case viewmodel.StepRetention:
		v.viewModel.SetRetention(value)
	}
}

func (v *AddTopicView) clearAndRender() {
	inputView, err := v.gui.View(topicWizardInput)
	if err != nil {
		return
	}
	inputView.Clear()
	inputView.SetCursor(0, 0)
	_ = v.render()
}

func (v *AddTopicView) Destroy(g *gocui.Gui) error {
	g.Cursor = false
	_ = g.DeleteView(topicWizardInput)
	return nil
}
