package views

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jroimartin/gocui"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	produceMessageView = "produce_message_view"
	produceMessageInput = "produce_message_input"

	produceMessageWidth  = 80
	produceMessageHeight = 12
)

type produceMessageEditor struct {
	onEsc     func()
	onTab     func()
	onEnter   func()
	onCtrlS   func()
	view      *ProduceMessageView
}

func (e *produceMessageEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch key {
	case gocui.KeyEsc:
		if e.onEsc != nil {
			e.onEsc()
		}
		return
	case gocui.KeyTab:
		if e.onTab != nil {
			e.onTab()
		}
		return
	case gocui.KeyEnter:
		if e.onEnter != nil {
			e.onEnter()
		}
		return
	case gocui.KeyCtrlS:
		if e.onCtrlS != nil {
			e.onCtrlS()
		}
		return
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

type ProduceMessageView struct {
	viewModel *viewmodel.ProduceMessageViewModel
	gui       *gocui.Gui
	onClose   func()
}

func NewProduceMessageView(vm *viewmodel.ProduceMessageViewModel, onClose func()) *ProduceMessageView {
	return &ProduceMessageView{
		viewModel: vm,
		onClose:   onClose,
	}
}

func (v *ProduceMessageView) GetViewModel() *viewmodel.ProduceMessageViewModel {
	return v.viewModel
}

func (v *ProduceMessageView) Initialize(g *gocui.Gui) error {
	v.gui = g
	return v.render()
}

func (v *ProduceMessageView) render() error {
	maxX, maxY := v.gui.Size()

	x0 := (maxX - produceMessageWidth) / 2
	y0 := (maxY - produceMessageHeight) / 2
	x1 := x0 + produceMessageWidth
	y1 := y0 + produceMessageHeight

	// Main form view
	formView, err := v.gui.SetView(produceMessageView, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	formView.Title = " " + v.viewModel.GetTitle() + " "
	formView.Editable = false

	// Render form content
	v.renderForm(formView)

	// Input field for the active field
	inputView, err := v.gui.SetView(produceMessageInput, x0+2, y0+9, x1-2, y1-2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	inputView.Editable = true
	inputView.Editor = &produceMessageEditor{
		onEsc:   v.handleEsc,
		onTab:   v.handleTab,
		onEnter: v.handleEnter,
		onCtrlS: v.handleCtrlS,
		view:    v,
	}

	v.updateInputView()

	_, _ = v.gui.SetViewOnTop(produceMessageView)
	_, _ = v.gui.SetViewOnTop(produceMessageInput)

	if _, err := v.gui.SetCurrentView(produceMessageInput); err != nil {
		slog.Error("failed to set current view", "view", produceMessageInput, "error", err)
	}

	v.gui.Cursor = true

	return nil
}

func (v *ProduceMessageView) renderForm(formView *gocui.View) {
	formView.Clear()

	topic := v.viewModel.GetTopic()
	key := v.viewModel.GetKey()
	value := v.viewModel.GetValue()
	currentField := v.viewModel.GetCurrentField()

	// Topic (readonly)
	topicPrefix := "  "
	if currentField == viewmodel.FieldTopic {
		topicPrefix = "> "
	}
	fmt.Fprintf(formView, "%sTopic: %s\n\n", topicPrefix, topic)

	// Key field
	keyPrefix := "  "
	if currentField == viewmodel.FieldKey {
		keyPrefix = "> "
	}
	keyDisplay := key
	if len(keyDisplay) > 60 {
		keyDisplay = keyDisplay[:57] + "..."
	}
	fmt.Fprintf(formView, "%sKey: %s\n\n", keyPrefix, keyDisplay)

	// Value field
	valuePrefix := "  "
	if currentField == viewmodel.FieldValue {
		valuePrefix = "> "
	}
	valueDisplay := value
	if len(valueDisplay) > 60 {
		valueDisplay = valueDisplay[:57] + "..."
	}
	fmt.Fprintf(formView, "%sValue: %s\n\n", valuePrefix, valueDisplay)

	// Validation errors
	if currentField == viewmodel.FieldTopic {
		if err := v.viewModel.GetValidationError(viewmodel.FieldTopic); err != "" {
			fmt.Fprintf(formView, "  [ERROR] %s\n", err)
		}
	}
	if currentField == viewmodel.FieldKey {
		if err := v.viewModel.GetValidationError(viewmodel.FieldKey); err != "" {
			fmt.Fprintf(formView, "  [ERROR] %s\n", err)
		}
	}
	if currentField == viewmodel.FieldValue {
		if err := v.viewModel.GetValidationError(viewmodel.FieldValue); err != "" {
			fmt.Fprintf(formView, "  [ERROR] %s\n", err)
		}
	}

	// Instructions
	fmt.Fprintln(formView, "  Tab: Next field | Esc: Cancel | Ctrl+S: Send")
}

func (v *ProduceMessageView) updateInputView() {
	inputView, err := v.gui.View(produceMessageInput)
	if err != nil {
		return
	}

	currentField := v.viewModel.GetCurrentField()
	fieldName := v.viewModel.GetFieldName(currentField)
	inputView.Title = " " + fieldName + " "

	inputView.Clear()
	inputView.SetCursor(0, 0)

	switch currentField {
	case viewmodel.FieldKey:
		if inputView.Buffer() == "" {
			fmt.Fprint(inputView, v.viewModel.GetKey())
		}
	case viewmodel.FieldValue:
		if inputView.Buffer() == "" {
			fmt.Fprint(inputView, v.viewModel.GetValue())
		}
	}
}

func (v *ProduceMessageView) handleEsc() {
	v.viewModel.Cancel()
}

func (v *ProduceMessageView) handleTab() {
	v.saveCurrentValue()
	v.viewModel.NextField()
	v.refreshForm()
	v.updateInputView()
}

func (v *ProduceMessageView) handleEnter() {
	// Enter moves to next field
	v.handleTab()
}

func (v *ProduceMessageView) saveCurrentValue() {
	inputView, err := v.gui.View(produceMessageInput)
	if err != nil {
		return
	}
	value := strings.TrimSpace(inputView.Buffer())

	currentField := v.viewModel.GetCurrentField()
	switch currentField {
	case viewmodel.FieldKey:
		v.viewModel.SetKey(value)
	case viewmodel.FieldValue:
		v.viewModel.SetValue(value)
	}
}

func (v *ProduceMessageView) refreshForm() {
	formView, err := v.gui.View(produceMessageView)
	if err != nil {
		return
	}
	v.renderForm(formView)
}

func (v *ProduceMessageView) Destroy(g *gocui.Gui) error {
	g.Cursor = false
	_ = g.DeleteView(produceMessageView)
	_ = g.DeleteView(produceMessageInput)
	return nil
}

func (v *ProduceMessageView) handleCtrlS() {
	if err := v.Submit(); err != nil {
		slog.Error("failed to send message", "error", err)
		v.showError(err.Error())
	}
}

func (v *ProduceMessageView) showError(msg string) {
	formView, err := v.gui.View(produceMessageView)
	if err != nil {
		return
	}
	v.renderForm(formView)
	fmt.Fprintf(formView, "\n  [ERROR] %s", msg)
}

func (v *ProduceMessageView) Submit() error {
	v.saveCurrentValue()

	if err := v.viewModel.Submit(); err != nil {
		return err
	}

	if v.onClose != nil {
		v.onClose()
	}
	return nil
}
