package tui

import (
	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
	"github.com/jurabek/lazykafka/internal/tui/views"
)

type PopupManager struct {
	gui                 *gocui.Gui
	layout              *Layout
	addBrokerView       *views.AddBrokerView
	addBrokerVM         *viewmodel.AddBrokerViewModel
	addTopicView        *views.AddTopicView
	addTopicVM          *viewmodel.AddTopicViewModel
	produceMessageView  *views.ProduceMessageView
	produceMessageVM    *viewmodel.ProduceMessageViewModel
	confirmView         *views.ConfirmView
	confirmVM           *viewmodel.ConfirmViewModel
	topicConfigView     *views.TopicConfigView
	topicConfigVM       *viewmodel.TopicConfigViewModel
	isPopupActive       bool
	activePopupView     string
	previousView        string
	onBrokerAdded       func(config models.BrokerConfig)
	onTopicAdded        func(config models.TopicConfig)
	onMessageProduced   func(topic string, key, value string, headers []models.Header) error
	onTopicConfigUpdate func(config models.TopicConfig)
}

func NewPopupManager(g *gocui.Gui, layout *Layout, onBrokerAdded func(models.BrokerConfig), onTopicAdded func(models.TopicConfig)) *PopupManager {
	return &PopupManager{
		gui:           g,
		layout:        layout,
		onBrokerAdded: onBrokerAdded,
		onTopicAdded:  onTopicAdded,
	}
}

func (pm *PopupManager) SetOnMessageProduced(fn func(topic string, key, value string, headers []models.Header) error) {
	pm.onMessageProduced = fn
}

func (pm *PopupManager) SetOnTopicConfigUpdate(fn func(config models.TopicConfig)) {
	pm.onTopicConfigUpdate = fn
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

func (pm *PopupManager) ShowProduceMessagePopup(topic string) error {
	if pm.isPopupActive {
		return nil
	}

	currentView := pm.gui.CurrentView()
	if currentView != nil {
		pm.previousView = currentView.Name()
	}

	pm.produceMessageVM = viewmodel.NewProduceMessageViewModel(
		func(t string, key, value string, headers []models.Header) error {
			if pm.onMessageProduced != nil {
				if err := pm.onMessageProduced(t, key, value, headers); err != nil {
					return err
				}
			}
			pm.Close()
			return nil
		},
		func() {
			pm.Close()
		},
	)
	pm.produceMessageVM.SetTopic(topic)

	pm.produceMessageView = views.NewProduceMessageView(pm.produceMessageVM, pm.Close)
	pm.isPopupActive = true
	pm.activePopupView = "produce_message_view"

	if err := pm.produceMessageView.Initialize(pm.gui); err != nil {
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

	if pm.produceMessageView != nil {
		_ = pm.produceMessageView.Destroy(pm.gui)
		pm.produceMessageView = nil
	}

	if pm.confirmView != nil {
		_ = pm.confirmView.Destroy(pm.gui)
		pm.confirmView = nil
	}

	if pm.topicConfigView != nil {
		_ = pm.topicConfigView.Destroy(pm.gui)
		pm.topicConfigView = nil
	}

	pm.addBrokerVM = nil
	pm.addTopicVM = nil
	pm.produceMessageVM = nil
	pm.confirmVM = nil
	pm.topicConfigVM = nil
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

func (pm *PopupManager) ShowConfirmPopup(message string, onYes func()) error {
	if pm.isPopupActive {
		return nil
	}

	currentView := pm.gui.CurrentView()
	if currentView != nil {
		pm.previousView = currentView.Name()
	}

	pm.confirmVM = viewmodel.NewConfirmViewModel(
		message,
		onYes,
		func() {
			pm.Close()
		},
	)

	pm.confirmView = views.NewConfirmView(pm.confirmVM, pm.Close)
	pm.isPopupActive = true
	pm.activePopupView = "confirm"

	if err := pm.confirmView.Initialize(pm.gui); err != nil {
		pm.isPopupActive = false
		pm.activePopupView = ""
		return err
	}

	return nil
}

func (pm *PopupManager) ShowTopicConfigPopup(topicName string, config models.TopicConfig) error {
	if pm.isPopupActive {
		return nil
	}

	currentView := pm.gui.CurrentView()
	if currentView != nil {
		pm.previousView = currentView.Name()
	}

	pm.topicConfigVM = viewmodel.NewTopicConfigViewModel(
		topicName,
		config,
		func(newConfig models.TopicConfig) {
			if pm.onTopicConfigUpdate != nil {
				pm.onTopicConfigUpdate(newConfig)
			}
			pm.Close()
		},
		func() {
			pm.Close()
		},
	)

	pm.topicConfigView = views.NewTopicConfigView(pm.topicConfigVM, func(newConfig models.TopicConfig) {
		if pm.onTopicConfigUpdate != nil {
			pm.onTopicConfigUpdate(newConfig)
		}
		pm.Close()
	}, func() {
		pm.Close()
	})

	pm.isPopupActive = true
	pm.activePopupView = "topic_config"

	if err := pm.topicConfigView.Initialize(pm.gui); err != nil {
		pm.isPopupActive = false
		pm.activePopupView = ""
		return err
	}

	return nil
}
