package tui

import (
	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
	"github.com/jurabek/lazykafka/internal/tui/views"
)

type PopupManager struct {
	gui             *gocui.Gui
	layout          *Layout
	addBrokerView   *views.AddBrokerView
	addBrokerVM     *viewmodel.AddBrokerViewModel
	addTopicView    *views.AddTopicView
	addTopicVM      *viewmodel.AddTopicViewModel
	isPopupActive   bool
	activePopupView string
	previousView    string
	onBrokerAdded   func(config models.BrokerConfig)
	onTopicAdded    func(config models.TopicConfig)
}

func NewPopupManager(g *gocui.Gui, layout *Layout, onBrokerAdded func(models.BrokerConfig), onTopicAdded func(models.TopicConfig)) *PopupManager {
	return &PopupManager{
		gui:           g,
		layout:        layout,
		onBrokerAdded: onBrokerAdded,
		onTopicAdded:  onTopicAdded,
	}
}

func (pm *PopupManager) IsActive() bool {
	return pm.isPopupActive
}

func (pm *PopupManager) ShowAddBrokerPopup() error {
	if pm.isPopupActive {
		return nil
	}

	currentView := pm.gui.CurrentView()
	if currentView != nil {
		pm.previousView = currentView.Name()
	}

	pm.addBrokerVM = viewmodel.NewAddBrokerViewModel(
		func(config models.BrokerConfig) {
			if pm.onBrokerAdded != nil {
				pm.onBrokerAdded(config)
			}
			pm.Close()
		},
		func() {
			pm.Close()
		},
	)

	pm.addBrokerView = views.NewAddBrokerView(pm.addBrokerVM, pm.Close)
	pm.isPopupActive = true
	pm.activePopupView = "wizard_input"

	if err := pm.addBrokerView.Initialize(pm.gui); err != nil {
		pm.isPopupActive = false
		pm.activePopupView = ""
		return err
	}

	return nil
}

func (pm *PopupManager) ShowAddTopicPopup() error {
	if pm.isPopupActive {
		return nil
	}

	currentView := pm.gui.CurrentView()
	if currentView != nil {
		pm.previousView = currentView.Name()
	}

	pm.addTopicVM = viewmodel.NewAddTopicViewModel(
		func(config models.TopicConfig) {
			if pm.onTopicAdded != nil {
				pm.onTopicAdded(config)
			}
			pm.Close()
		},
		func() {
			pm.Close()
		},
	)

	pm.addTopicView = views.NewAddTopicView(pm.addTopicVM, pm.Close)
	pm.isPopupActive = true
	pm.activePopupView = "topic_wizard_input"

	if err := pm.addTopicView.Initialize(pm.gui); err != nil {
		pm.isPopupActive = false
		pm.activePopupView = ""
		return err
	}

	return nil
}

func (pm *PopupManager) Close() {
	if !pm.isPopupActive {
		return
	}

	if pm.addBrokerView != nil {
		_ = pm.addBrokerView.Destroy(pm.gui)
		pm.addBrokerView = nil
	}

	if pm.addTopicView != nil {
		_ = pm.addTopicView.Destroy(pm.gui)
		pm.addTopicView = nil
	}

	pm.addBrokerVM = nil
	pm.addTopicVM = nil
	pm.isPopupActive = false
	pm.activePopupView = ""

	if pm.previousView != "" {
		_, _ = pm.gui.SetCurrentView(pm.previousView)
	}
}

func (pm *PopupManager) BringToTop() {
	if !pm.isPopupActive || pm.activePopupView == "" {
		return
	}
	if v, err := pm.gui.View(pm.activePopupView); err == nil {
		pm.gui.SetViewOnTop(v.Name())
		_, _ = pm.gui.SetCurrentView(pm.activePopupView)
	}
}
