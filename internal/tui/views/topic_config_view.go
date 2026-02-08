package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

type TopicConfigView struct {
	viewModel *viewmodel.TopicConfigViewModel
	onSave    func(models.TopicConfig)
	onClose   func()
	inputView string
	gui       *gocui.Gui
}

func NewTopicConfigView(vm *viewmodel.TopicConfigViewModel, onSave func(models.TopicConfig), onClose func()) *TopicConfigView {
	return &TopicConfigView{
		viewModel: vm,
		onSave:    onSave,
		onClose:   onClose,
		inputView: "topic_config_input",
	}
}

func (v *TopicConfigView) Initialize(g *gocui.Gui) error {
	v.gui = g
	maxX, maxY := g.Size()
	width := 60
	height := 14
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2
	x1 := x0 + width
	y1 := y0 + height

	viewName := v.viewModel.GetName()
	view, err := g.SetView(viewName, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = v.viewModel.GetTitle()
	view.Editable = false

	if _, err := g.SetView(v.inputView, x0+2, y0+height-2, x1-2, y0+height-1); err != nil && err != gocui.ErrUnknownView {
		return err
	}

	inputView, _ := g.View(v.inputView)
	inputView.Editable = true
	inputView.Frame = false
	inputView.Editable = true

	editor := &topicConfigEditor{view: v}

	if err := g.SetKeybinding(v.inputView, gocui.KeyCtrlS, gocui.ModNone, editor.handleSave); err != nil {
		return err
	}
	if err := g.SetKeybinding(v.inputView, gocui.KeyEsc, gocui.ModNone, editor.handleEsc); err != nil {
		return err
	}
	if err := g.SetKeybinding(v.inputView, gocui.KeyTab, gocui.ModNone, editor.handleTab); err != nil {
		return err
	}
	if err := g.SetKeybinding(v.inputView, gocui.KeyEnter, gocui.ModNone, editor.handleEnter); err != nil {
		return err
	}

	v.updateInputForField()
	g.SetCurrentView(v.inputView)

	return nil
}

func (v *TopicConfigView) updateInputForField() {
	if v.gui == nil {
		return
	}

	v.gui.Update(func(gui *gocui.Gui) error {
		configView, _ := gui.View(v.viewModel.GetName())
		if configView != nil {
			v.Render(gui, configView)
		}

		inputView, _ := gui.View(v.inputView)
		if inputView != nil {
			inputView.Clear()
			inputView.SetCursor(0, 0)
			config := v.viewModel.GetConfig()
			field := v.viewModel.GetCurrentField()

			var value string
			switch field {
			case viewmodel.FieldConfigPartitions:
				value = strconv.Itoa(config.Partitions)
			case viewmodel.FieldConfigReplication:
				value = strconv.Itoa(config.ReplicationFactor)
			case viewmodel.FieldConfigCleanup:
				value = config.CleanupPolicy.String()
			case viewmodel.FieldConfigMinSync:
				value = strconv.Itoa(config.MinInSyncReplicas)
			case viewmodel.FieldConfigRetention:
				value = formatRetention(config.RetentionMs)
			}

			fmt.Fprint(inputView, value)
		}

		_, _ = gui.SetCurrentView(v.inputView)
		return nil
	})
}

func formatRetention(ms int64) string {
	if ms <= 0 {
		return ""
	}
	const millsPerDay = 24 * int64(time.Hour/time.Millisecond)
	days := ms / millsPerDay
	if days > 0 && ms%millsPerDay == 0 {
		return strconv.FormatInt(days, 10) + "d"
	}
	return (time.Duration(ms) * time.Millisecond).String()
}

func parseRetention(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}

	const millsPerDay = 24 * int64(time.Hour/time.Millisecond)
	if daysStr, ok := strings.CutSuffix(s, "d"); ok {
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0
		}
		return int64(days) * millsPerDay
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d.Milliseconds()
}

func (v *TopicConfigView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	config := v.viewModel.GetConfig()
	currentField := v.viewModel.GetCurrentField()

	fields := []struct {
		label    string
		value    string
		readonly bool
		fieldNum int
	}{
		{"Partitions", strconv.Itoa(config.Partitions), true, viewmodel.FieldConfigPartitions},
		{"Replication Factor", strconv.Itoa(config.ReplicationFactor), false, viewmodel.FieldConfigReplication},
		{"Cleanup Policy", config.CleanupPolicy.String(), false, viewmodel.FieldConfigCleanup},
		{"Min In-Sync Replicas", strconv.Itoa(config.MinInSyncReplicas), false, viewmodel.FieldConfigMinSync},
		{"Retention", formatRetention(config.RetentionMs), false, viewmodel.FieldConfigRetention},
	}

	for _, f := range fields {
		prefix := "  "
		if f.fieldNum == currentField {
			prefix = "> "
		}
		fmt.Fprintf(gocuiView, "%s%s: %s\n", prefix, f.label, f.value)

		if f.fieldNum == currentField {
			labelLower := strings.ToLower(f.label)
			if errMsg := v.viewModel.GetFieldError(labelLower); errMsg != "" {
				fmt.Fprintf(gocuiView, "    Error: %s\n", errMsg)
			}
		}
	}

	fmt.Fprintln(gocuiView, "\n  Tab: next field | Ctrl+S: save | Esc: cancel")

	return nil
}

func (v *TopicConfigView) Destroy(g *gocui.Gui) error {
	_ = g.DeleteView(v.viewModel.GetName())
	_ = g.DeleteView(v.inputView)
	g.DeleteKeybindings(v.viewModel.GetName())
	g.DeleteKeybindings(v.inputView)
	return nil
}

type topicConfigEditor struct {
	view *TopicConfigView
}

func (e *topicConfigEditor) handleSave(g *gocui.Gui, v *gocui.View) error {
	// Get current value from input and update config
	inputView, _ := g.View(e.view.inputView)
	if inputView == nil {
		return nil
	}

	buffer := inputView.ViewBuffer()
	value := strings.TrimSpace(buffer)
	config := e.view.viewModel.GetConfig()
	field := e.view.viewModel.GetCurrentField()

	var err error
	switch field {
	case viewmodel.FieldConfigPartitions:
		// readonly - ignore
	case viewmodel.FieldConfigReplication:
		config.ReplicationFactor, err = strconv.Atoi(value)
	case viewmodel.FieldConfigCleanup:
		switch strings.ToLower(value) {
		case "delete":
			config.CleanupPolicy = 0
		case "compact":
			config.CleanupPolicy = 1
		case "compact,delete":
			config.CleanupPolicy = 2
		default:
			e.view.viewModel.SetFieldError("cleanup policy", "must be: delete, compact, or compact,delete")
			e.view.updateInputForField()
			return nil
		}
	case viewmodel.FieldConfigMinSync:
		config.MinInSyncReplicas, err = strconv.Atoi(value)
	case viewmodel.FieldConfigRetention:
		config.RetentionMs = parseRetention(value)
	}

	if err != nil {
		e.view.viewModel.SetFieldError("value", "invalid format")
		e.view.updateInputForField()
		return nil
	}

	e.view.viewModel.SetConfig(config)

	if validateErr := e.view.viewModel.Validate(); validateErr != nil {
		e.view.updateInputForField()
		return nil
	}

	if e.view.onSave != nil {
		e.view.onSave(config)
	}

	return nil
}

func (e *topicConfigEditor) handleEsc(g *gocui.Gui, v *gocui.View) error {
	if e.view.onClose != nil {
		e.view.onClose()
	}
	return nil
}

func (e *topicConfigEditor) handleTab(g *gocui.Gui, v *gocui.View) error {
	// Save current field first
	e.handleSave(g, v)
	e.view.viewModel.NextField()
	e.view.updateInputForField()
	return nil
}

func (e *topicConfigEditor) handleEnter(g *gocui.Gui, v *gocui.View) error {
	return e.handleTab(g, v)
}
