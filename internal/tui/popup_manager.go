package tui

import (
	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	viewmodel "github.com/jurabek/lazykafka/internal/tui/view_models"
	"github.com/jurabek/lazykafka/internal/tui/views"
)

type PopupManager struct {
	gui           *gocui.Gui
	layout        *Layout
	addBrokerView *views.AddBrokerView
	addBrokerVM   *viewmodel.AddBrokerViewModel
	isPopupActive bool
	previousView  string
	onBrokerAdded func(config models.BrokerConfig)
}

func NewPopupManager(g *gocui.Gui, layout *Layout, onBrokerAdded func(models.BrokerConfig)) *PopupManager {
	return &PopupManager{
		gui:           g,
		layout:        layout,
		onBrokerAdded: onBrokerAdded,
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

	if err := pm.addBrokerView.Initialize(pm.gui); err != nil {
		pm.isPopupActive = false
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

	pm.addBrokerVM = nil
	pm.isPopupActive = false

	if pm.previousView != "" {
		_, _ = pm.gui.SetCurrentView(pm.previousView)
	}
}
