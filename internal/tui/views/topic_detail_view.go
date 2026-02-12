package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	messageDetailPopup = "message_detail_popup"
	filterPopupView    = "filter_popup"
	filterPopupInput   = "filter_popup_input"
)

type TopicDetailView struct {
	BaseView
	viewModel         *viewmodel.TopicDetailViewModel
	messageBrowserVM  *viewmodel.MessageBrowserViewModel
	messageBrowser    *MessageBrowserView
	messageDetailVM   *viewmodel.MessageDetailViewModel
	messageDetailView *MessageDetailView
	gui               *gocui.Gui
	currentTab        viewmodel.TabType
}

func NewTopicDetailView(vm *viewmodel.TopicDetailViewModel, messageBrowserVM *viewmodel.MessageBrowserViewModel) *TopicDetailView {
	return &TopicDetailView{
		BaseView:         BaseView{viewModel: vm},
		viewModel:        vm,
		messageBrowserVM: messageBrowserVM,
		currentTab:       viewmodel.TabPartitions,
	}
}

func (v *TopicDetailView) Initialize(g *gocui.Gui) (bool, error) {
	v.gui = g
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

	// Initialize message browser view with separate view name
	if v.messageBrowser == nil {
		v.messageBrowser = NewMessageBrowserView(v.messageBrowserVM)
		_, err := g.SetView(v.messageBrowserVM.GetName(), x0, y0, x1, y1)
		if err != nil && err != gocui.ErrUnknownView {
			return created, err
		}
		v.messageBrowser.SetBounds(x0, y0, x1, y1)
		if _, err := v.messageBrowser.Initialize(g); err != nil {
			return created, err
		}
	}

	// Initialize message detail view as popup
	if v.messageDetailVM == nil {
		v.messageDetailVM = viewmodel.NewMessageDetailViewModel()
		v.messageDetailView = NewMessageDetailView(v.messageDetailVM)
		v.messageDetailView.SetupCallbacks(g)

		// Set up close callback for detail view
		v.messageDetailVM.SetOnClose(func() {
			v.closeMessageDetail()
		})
	}

	return created, nil
}

func (v *TopicDetailView) Render(g *gocui.Gui, gocuiView *gocui.View) error {
	gocuiView.Clear()
	gocuiView.Title = v.viewModel.GetTitle()

	activeTab := v.viewModel.GetActiveTab()
	tabChanged := activeTab != v.currentTab
	v.currentTab = activeTab

	if activeTab == viewmodel.TabMessages {
		// Move message browser view to top
		if msgView, err := g.View(v.messageBrowserVM.GetName()); err == nil {
			g.SetViewOnTop(msgView.Name())
			if tabChanged {
				return v.messageBrowser.Render(g, msgView)
			}
		}
		// Move main detail view to bottom
		g.SetViewOnBottom(gocuiView.Name())
	} else {
		// Move main detail view to top
		g.SetViewOnTop(gocuiView.Name())
		// Move message browser view to bottom
		if msgView, err := g.View(v.messageBrowserVM.GetName()); err == nil {
			g.SetViewOnBottom(msgView.Name())
		}

		maxX, _ := gocuiView.Size()
		content := v.viewModel.RenderPartitionsTable(maxX)
		fmt.Fprint(gocuiView, content)
	}

	return nil
}

func (v *TopicDetailView) Destroy(g *gocui.Gui) error {
	_ = g.DeleteView(v.viewModel.GetName())
	_ = g.DeleteView(v.messageBrowserVM.GetName())
	_ = g.DeleteView(messageDetailPopup)
	if v.messageBrowser != nil {
		_ = v.messageBrowser.Destroy(g)
	}
	if v.messageDetailView != nil {
		_ = v.messageDetailView.Destroy(g)
	}
	return nil
}

func (v *TopicDetailView) SetupCallbacks(g *gocui.Gui) {
	v.viewModel.SetOnChange(func(event types.ChangeEvent) {
		g.Update(func(gui *gocui.Gui) error {
			view, err := g.View(v.viewModel.GetName())
			if err != nil {
				return nil
			}

			// Update topic for message browser when tab changes to Messages
			if event.FieldName == "tab" || event.FieldName == types.FieldSelectedIndex {
				activeTab := v.viewModel.GetActiveTab()
				if activeTab == viewmodel.TabMessages {
					topic := v.viewModel.GetTopic()
					if topic != nil {
						v.messageBrowserVM.SetTopic(topic.Name)
						v.messageBrowserVM.LoadMessages(models.MessageFilter{
							Partition: -1,
							Offset:    -1,
							Limit:     100,
							Format:    "json",
						})
					}
				}
			}

			return v.Render(g, view)
		})
	})

	// Set up message selection callback to show detail view
	v.messageBrowserVM.SetOnMessageSelected(func(msg *models.Message) {
		if msg != nil {
			v.showMessageDetail(g, msg)
		}
	})
}

func (v *TopicDetailView) showMessageDetail(g *gocui.Gui, msg *models.Message) {
	g.Update(func(gui *gocui.Gui) error {
		// Get current terminal size
		maxX, maxY := g.Size()

		// Calculate popup dimensions (80% of screen, centered)
		popupWidth := maxX * 4 / 5
		popupHeight := maxY * 4 / 5
		if popupWidth < 60 {
			popupWidth = 60
		}
		if popupHeight < 15 {
			popupHeight = 15
		}

		x0 := (maxX - popupWidth) / 2
		y0 := (maxY - popupHeight) / 2
		x1 := x0 + popupWidth
		y1 := y0 + popupHeight

		// Create or update popup view
		detailView, err := g.SetView(messageDetailPopup, x0, y0, x1, y1)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}

		detailView.Title = " Message Details "
		detailView.Wrap = false
		detailView.Editable = false

		// Set the message and render
		v.messageDetailVM.SetMessage(msg)
		if err := v.messageDetailView.Render(g, detailView); err != nil {
			return err
		}

		// Set up keybinding for closing detail view
		_ = g.SetKeybinding(messageDetailPopup, 'q', gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
			v.closeMessageDetail()
			return nil
		})
		_ = g.SetKeybinding(messageDetailPopup, gocui.KeyEsc, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
			v.closeMessageDetail()
			return nil
		})

		// Bring detail view to top and set as current
		g.SetViewOnTop(messageDetailPopup)
		_, _ = g.SetCurrentView(messageDetailPopup)

		return nil
	})
}

func (v *TopicDetailView) closeMessageDetail() {
	if v.gui == nil {
		return
	}
	v.gui.Update(func(g *gocui.Gui) error {
		_ = g.DeleteView(messageDetailPopup)
		return nil
	})
}

func (v *TopicDetailView) ShowFilterPopup() {
	if v.gui == nil {
		return
	}

	v.messageBrowserVM.ShowFilterPopup()

	v.gui.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		width := 50
		height := 10
		x0 := (maxX - width) / 2
		y0 := (maxY - height) / 2
		x1 := x0 + width
		y1 := y0 + height

		popupView, err := g.SetView(filterPopupView, x0, y0, x1, y1)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		popupView.Title = " Filter Messages "
		popupView.Editable = false

		inputView, err := g.SetView(filterPopupInput, x0+2, y1-2, x1-2, y1-1)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		inputView.Editable = true
		inputView.Frame = false

		v.renderFilterPopup(g)
		v.setupFilterKeybindings(g)

		g.SetViewOnTop(filterPopupView)
		g.SetViewOnTop(filterPopupInput)
		_, _ = g.SetCurrentView(filterPopupInput)
		return nil
	})
}

func (v *TopicDetailView) renderFilterPopup(g *gocui.Gui) {
	popupView, err := g.View(filterPopupView)
	if err != nil {
		return
	}
	popupView.Clear()

	currentField := v.messageBrowserVM.GetCurrentFilterField()
	partition := v.messageBrowserVM.GetPendingPartition()
	offsetMode := v.messageBrowserVM.GetPendingOffsetMode()
	limit := v.messageBrowserVM.GetPendingLimit()

	partitionLabel := "all"
	if partition >= 0 {
		partitionLabel = fmt.Sprintf("%d", partition)
	}

	fields := []struct {
		label string
		value string
	}{
		{"Partition (-1=all)", partitionLabel},
		{"Offset Mode", offsetMode},
		{"Limit", fmt.Sprintf("%d", limit)},
	}

	for i, f := range fields {
		prefix := "  "
		if i == currentField {
			prefix = "> "
		}
		fmt.Fprintf(popupView, "%s%s: %s\n", prefix, f.label, f.value)
	}

	fmt.Fprintln(popupView, "\n  Tab: next | Enter: apply | Esc: cancel")

	// Update input view with current field value
	inputView, err := g.View(filterPopupInput)
	if err != nil {
		return
	}
	inputView.Clear()
	inputView.SetCursor(0, 0)

	switch currentField {
	case viewmodel.FilterFieldPartition:
		fmt.Fprint(inputView, fmt.Sprintf("%d", partition))
	case viewmodel.FilterFieldOffsetMode:
		fmt.Fprint(inputView, offsetMode)
	case viewmodel.FilterFieldLimit:
		fmt.Fprint(inputView, fmt.Sprintf("%d", limit))
	}
}

func (v *TopicDetailView) setupFilterKeybindings(g *gocui.Gui) {
	_ = g.SetKeybinding(filterPopupInput, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.closeFilterPopup()
		return nil
	})

	_ = g.SetKeybinding(filterPopupInput, gocui.KeyTab, gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.applyCurrentFilterField(g)
		v.messageBrowserVM.NextFilterField()
		v.renderFilterPopup(g)
		return nil
	})

	_ = g.SetKeybinding(filterPopupInput, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, view *gocui.View) error {
		v.applyCurrentFilterField(g)
		v.messageBrowserVM.ApplyFilter()
		v.closeFilterPopup()
		return nil
	})
}

func (v *TopicDetailView) applyCurrentFilterField(g *gocui.Gui) {
	inputView, err := g.View(filterPopupInput)
	if err != nil {
		return
	}

	value := strings.TrimSpace(inputView.ViewBuffer())
	currentField := v.messageBrowserVM.GetCurrentFilterField()

	switch currentField {
	case viewmodel.FilterFieldPartition:
		if val, err := strconv.Atoi(value); err == nil {
			v.messageBrowserVM.SetPendingPartition(val)
		}
	case viewmodel.FilterFieldOffsetMode:
		mode := strings.ToLower(value)
		if mode == viewmodel.OffsetModeNewest || mode == viewmodel.OffsetModeOldest {
			v.messageBrowserVM.SetPendingOffsetMode(mode)
		}
	case viewmodel.FilterFieldLimit:
		if val, err := strconv.Atoi(value); err == nil && val > 0 {
			v.messageBrowserVM.SetPendingLimit(val)
		}
	}
}

func (v *TopicDetailView) closeFilterPopup() {
	if v.gui == nil {
		return
	}
	v.messageBrowserVM.CloseFilterPopup()
	v.gui.Update(func(g *gocui.Gui) error {
		_ = g.DeleteView(filterPopupInput)
		_ = g.DeleteView(filterPopupView)
		g.DeleteKeybindings(filterPopupInput)

		// Update the message browser title to reflect current filter
		if msgView, err := g.View(v.messageBrowserVM.GetName()); err == nil {
			msgView.Title = v.messageBrowserVM.GetTitle()
		}

		return nil
	})
}
