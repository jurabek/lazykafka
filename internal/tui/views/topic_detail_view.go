package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	messageDetailPopup = "message_detail_popup"
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
		// Delete the popup view
		_ = g.DeleteView(messageDetailPopup)
		return nil
	})
}
