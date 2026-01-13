package tui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/data"
	"github.com/jurabek/lazykafka/internal/kafka"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/secrets"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
	"github.com/jurabek/lazykafka/internal/tui/views"
)

const sidebarWidth = 40

const (
	sidebarBrokers = iota
	sidebarTopics
	sidebarConsumerGroups
	sidebarSchemaRegistry
)

type Layout struct {
	sidebarViews      []views.View
	detailViews       map[int]views.View
	activeViewIndex   int
	activeDetailIndex int
	gui               *gocui.Gui
	mainVM            *viewmodel.MainViewModel
	popupManager      *PopupManager
	brokerStorage     data.BrokerStorage
	secretStore       secrets.SecretStore
	brokerConfigs     []models.BrokerConfig
	statusMessage     string
	statusMu          sync.RWMutex
	initialDataLoaded bool
}

func NewLayout(ctx context.Context, g *gocui.Gui) *Layout {
	brokerStorage, _ := data.NewFileBrokerStorage()

	var configs []models.BrokerConfig
	if brokerStorage != nil {
		configs, _ = brokerStorage.Load()
	}

	clientFactory := kafka.NewFranzClientFactory()
	mainVM := viewmodel.NewMainViewModel(ctx, configs, clientFactory)

	brokersView := views.NewBrokersView(mainVM.BrokersVM())
	topicsView := views.NewTopicsView(mainVM.TopicsVM())
	cgView := views.NewConsumerGroupsView(mainVM.ConsumerGroupsVM())
	srView := views.NewSchemaRegistryView(mainVM.SchemaRegistryVM())

	topicDetailView := views.NewTopicDetailView(mainVM.TopicDetailVM())
	cgDetailView := views.NewConsumerGroupDetailView(mainVM.ConsumerGroupDetailVM())
	srDetailView := views.NewSchemaRegistryDetailView(mainVM.SchemaRegistryDetailVM())

	sidebarViews := []views.View{brokersView, topicsView, cgView, srView}
	detailViews := map[int]views.View{
		sidebarTopics:         topicDetailView,
		sidebarConsumerGroups: cgDetailView,
		sidebarSchemaRegistry: srDetailView,
	}

	for i, v := range sidebarViews {
		v.SetActive(i == 0)
		v.SetupCallbacks(g)
	}
	for _, v := range detailViews {
		v.SetupCallbacks(g)
	}

	mainVM.BrokersVM().SetGui(g)

	secretStore := secrets.NewKeyringStore()

	layout := &Layout{
		sidebarViews:      sidebarViews,
		detailViews:       detailViews,
		activeViewIndex:   0,
		activeDetailIndex: -1,
		gui:               g,
		mainVM:            mainVM,
		brokerStorage:     brokerStorage,
		secretStore:       secretStore,
		brokerConfigs:     configs,
	}

	mainVM.SetOnError(func(err error) {
		layout.SetStatusMessage(err.Error())
	})

	layout.popupManager = NewPopupManager(g, layout, func(config models.BrokerConfig) {
		layout.onBrokerAdded(config)
	}, func(config models.TopicConfig) {
		layout.onTopicAdded(config)
	})

	return layout
}

func (l *Layout) SetStatusMessage(msg string) {
	l.statusMu.Lock()
	l.statusMessage = msg
	l.statusMu.Unlock()

	l.gui.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		helpHeight := 2
		if v, err := g.View(panelHelp); err == nil {
			l.renderHelpView(v, maxX, maxY, helpHeight)
		}
		return nil
	})
}

func (l *Layout) ClearStatusMessage() {
	l.SetStatusMessage("")
}

func (l *Layout) Manager(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if maxX < minTermWidth || maxY < minTermHeight {
		return fmt.Errorf("terminal too small: need at least %dx%d", minTermWidth, minTermHeight)
	}

	helpHeight := 2
	sidebarX := sidebarWidth
	if sidebarX > maxX/2 {
		sidebarX = maxX / 2
	}

	remainingViews := len(l.sidebarViews) - 1 // Exclude brokers from equal height calculation
	remainingHeight := maxY - helpHeight - 2 - maxBrokersHeight
	otherPanelHeight := remainingHeight / remainingViews

	brokersY1 := maxBrokersHeight - 1

	for i, view := range l.sidebarViews {
		if i == sidebarBrokers {
			// Brokers get fixed smaller height
			view.SetBounds(0, 0, sidebarX, brokersY1)
		} else {
			// Other views share the remaining height
			y0 := maxBrokersHeight + (i-1)*otherPanelHeight
			y1 := y0 + otherPanelHeight - 1
			if i == len(l.sidebarViews)-1 {
				y1 = maxY - helpHeight - 2
			}
			view.SetBounds(0, y0, sidebarX, y1)
		}

		created, err := view.Initialize(g)
		if err != nil {
			return fmt.Errorf("initializing view %s: %w", view.GetViewModel().GetName(), err)
		}

		if created {
			if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
				if err := view.Render(g, gocuiView); err != nil {
					return fmt.Errorf("rendering view: %w", err)
				}
			}
		}
	}

	if err := l.manageDetailViews(g, sidebarX+1, 0, maxX-1, maxY-helpHeight-2); err != nil {
		return err
	}

	if err := l.createHelpView(g, maxX, maxY, helpHeight); err != nil {
		return fmt.Errorf("creating help view: %w", err)
	}

	if l.popupManager == nil || !l.popupManager.IsActive() {
		activeView := l.sidebarViews[l.activeViewIndex]
		if _, err := g.SetCurrentView(activeView.GetViewModel().GetName()); err != nil {
			return fmt.Errorf("setting current view: %w", err)
		}
	}

	if !l.initialDataLoaded {
		l.initialDataLoaded = true
		l.mainVM.LoadInitialData()
	}

	if l.popupManager != nil {
		l.popupManager.BringToTop()
	}

	return nil
}

func (l *Layout) manageDetailViews(g *gocui.Gui, x0, y0, x1, y1 int) error {
	_, hasDetail := l.detailViews[l.activeViewIndex]

	for idx, view := range l.detailViews {
		if idx == l.activeViewIndex && hasDetail {
			view.SetBounds(x0, y0, x1, y1)
			created, err := view.Initialize(g)
			if err != nil {
				return fmt.Errorf("initializing detail view: %w", err)
			}
			if created || l.activeDetailIndex != idx {
				if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
					if err := view.Render(g, gocuiView); err != nil {
						return fmt.Errorf("rendering detail view: %w", err)
					}
				}
			}
			if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
				g.SetViewOnTop(gocuiView.Name())
			}
		} else {
			if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
				g.SetViewOnBottom(gocuiView.Name())
			}
		}
	}

	if hasDetail {
		l.activeDetailIndex = l.activeViewIndex
	} else {
		l.activeDetailIndex = -1
	}

	return nil
}

func (l *Layout) createHelpView(g *gocui.Gui, maxX, maxY, helpHeight int) error {
	v, err := g.SetView(panelHelp, 0, maxY-helpHeight-1, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Frame = false
	l.renderHelpView(v, maxX, maxY, helpHeight)
	return nil
}

func (l *Layout) renderHelpView(v *gocui.View, maxX, maxY, helpHeight int) {
	v.Clear()

	l.statusMu.RLock()
	statusMsg := l.statusMessage
	l.statusMu.RUnlock()

	if statusMsg != "" {
		fmt.Fprintf(v, " Error: %s\n", statusMsg)
	} else {
		fmt.Fprintln(v, " ←/→: switch panel | ↑/k: up | ↓/j: down | 1-4: jump panel | n: new | e: edit config | q: quit")
	}
}

func (l *Layout) NextPanel(g *gocui.Gui) {
	l.activeViewIndex = (l.activeViewIndex + 1) % len(l.sidebarViews)
	l.refreshAllViews(g)
}

func (l *Layout) PrevPanel(g *gocui.Gui) {
	l.activeViewIndex--
	if l.activeViewIndex < 0 {
		l.activeViewIndex = len(l.sidebarViews) - 1
	}
	l.refreshAllViews(g)
}

func (l *Layout) JumpToPanel(g *gocui.Gui, index int) {
	if index >= 0 && index < len(l.sidebarViews) {
		l.activeViewIndex = index
		l.refreshAllViews(g)
	}
}

func (l *Layout) refreshAllViews(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		for i, view := range l.sidebarViews {
			view.SetActive(i == l.activeViewIndex)
			if gocuiView, err := g.View(view.GetViewModel().GetName()); err == nil {
				view.Render(g, gocuiView)
			}
		}
		activeView := l.sidebarViews[l.activeViewIndex]
		_, _ = g.SetCurrentView(activeView.GetViewModel().GetName())
		return nil
	})
}

func (l *Layout) MainViewModel() *viewmodel.MainViewModel {
	return l.mainVM
}

func (l *Layout) ShowAddBrokerPopup() error {
	return l.popupManager.ShowAddBrokerPopup()
}

func (l *Layout) ShowAddTopicPopup() error {
	return l.popupManager.ShowAddTopicPopup()
}

func (l *Layout) GetActiveViewIndex() int {
	return l.activeViewIndex
}

func (l *Layout) IsPopupActive() bool {
	return l.popupManager.IsActive()
}

func (l *Layout) ClosePopup() {
	l.popupManager.Close()
}

func (l *Layout) onBrokerAdded(config models.BrokerConfig) {
	l.mainVM.BrokersVM().AddBrokerConfig(config)
	l.mainVM.AddBrokerConfig(config)
	l.brokerConfigs = append(l.brokerConfigs, config)

	if config.AuthType == models.AuthSASL {
		if l.secretStore != nil && config.Password != "" {
			_ = l.secretStore.SaveCredentials(config.Name, config.Username, config.Password)
		}
	}

	if l.brokerStorage != nil {
		configToSave := config
		configToSave.Password = ""

		configs, _ := l.brokerStorage.Load()
		configs = append(configs, configToSave)
		_ = l.brokerStorage.Save(configs)
	}

	l.gui.Update(func(g *gocui.Gui) error {
		if view, err := g.View(panelBrokers); err == nil {
			brokersView := l.sidebarViews[sidebarBrokers]
			_ = brokersView.Render(g, view)
		}
		return nil
	})
}

func (l *Layout) onTopicAdded(config models.TopicConfig) {
	ctx := context.Background()
	if err := l.mainVM.CreateTopic(ctx, config); err != nil {
		slog.Error("creating topic failed", slog.Any("error", err))
		l.SetStatusMessage(err.Error())
		return
	}

	l.mainVM.TopicsVM().Reload()

	l.gui.Update(func(g *gocui.Gui) error {
		if view, err := g.View(panelTopics); err == nil {
			topicsView := l.sidebarViews[sidebarTopics]
			_ = topicsView.Render(g, view)
		}
		return nil
	})
}
