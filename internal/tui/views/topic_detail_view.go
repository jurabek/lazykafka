package views

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
)

const (
	topicDetailMessagesView = "topic_detail_messages"
)

type TopicDetailView struct {
	BaseView
	viewModel        *viewmodel.TopicDetailViewModel
	messageBrowserVM *viewmodel.MessageBrowserViewModel
	messageBrowser   *MessageBrowserView
	gui              *gocui.Gui
	currentTab       viewmodel.TabType
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
	if v.messageBrowser != nil {
		_ = v.messageBrowser.Destroy(g)
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
}
