package tui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
)

type Panel struct {
	Name          string
	Title         string
	Items         []string
	SelectedIndex int
}

type Layout struct {
	panels            []*Panel
	activePanelIndex  int
	brokers           []models.Broker
	topics            []models.Topic
	consumerGroups    []models.ConsumerGroup
	schemaRegistries  []models.SchemaRegistry
}

func NewLayout() *Layout {
	brokers := models.MockBrokers()
	topics := models.MockTopics()
	consumerGroups := models.MockConsumerGroups()
	schemaRegistries := models.MockSchemaRegistries()

	brokerItems := make([]string, len(brokers))
	for i, b := range brokers {
		brokerItems[i] = fmt.Sprintf("%d: %s:%d", b.ID, b.Host, b.Port)
	}

	topicItems := make([]string, len(topics))
	for i, t := range topics {
		topicItems[i] = fmt.Sprintf("%s (P:%d R:%d)", t.Name, t.Partitions, t.Replicas)
	}

	cgItems := make([]string, len(consumerGroups))
	for i, cg := range consumerGroups {
		cgItems[i] = fmt.Sprintf("%s [%s] members:%d", cg.Name, cg.State, cg.Members)
	}

	srItems := make([]string, len(schemaRegistries))
	for i, sr := range schemaRegistries {
		srItems[i] = fmt.Sprintf("%s v%d (%s)", sr.Subject, sr.Version, sr.Type)
	}

	return &Layout{
		panels: []*Panel{
			{Name: panelBrokers, Title: "Brokers", Items: brokerItems, SelectedIndex: 0},
			{Name: panelTopics, Title: "Topics", Items: topicItems, SelectedIndex: 0},
			{Name: panelConsumerGroups, Title: "Consumer Groups", Items: cgItems, SelectedIndex: 0},
			{Name: panelSchemaRegistry, Title: "Schema Registry", Items: srItems, SelectedIndex: 0},
		},
		activePanelIndex: 0,
		brokers:          brokers,
		topics:           topics,
		consumerGroups:   consumerGroups,
		schemaRegistries: schemaRegistries,
	}
}

func (l *Layout) Manager(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if maxX < minTermWidth || maxY < minTermHeight {
		return fmt.Errorf("terminal too small: need at least %dx%d", minTermWidth, minTermHeight)
	}

	panelHeight := (maxY - 3) / len(l.panels)
	helpHeight := 2

	for i, panel := range l.panels {
		y0 := i * panelHeight
		y1 := y0 + panelHeight - 1
		if i == len(l.panels)-1 {
			y1 = maxY - helpHeight - 2
		}

		if err := l.createPanelView(g, panel, 0, y0, maxX-1, y1, i == l.activePanelIndex); err != nil {
			return fmt.Errorf("creating panel %s: %w", panel.Name, err)
		}
	}

	if err := l.createHelpView(g, maxX, maxY, helpHeight); err != nil {
		return fmt.Errorf("creating help view: %w", err)
	}

	activePanel := l.panels[l.activePanelIndex]
	if _, err := g.SetCurrentView(activePanel.Name); err != nil {
		return fmt.Errorf("setting current view: %w", err)
	}

	return nil
}

func (l *Layout) createPanelView(g *gocui.Gui, panel *Panel, x0, y0, x1, y1 int, _ bool) error {
	v, err := g.SetView(panel.Name, x0, y0, x1, y1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	panelNum := l.getPanelIndex(panel.Name) + 1
	v.Title = fmt.Sprintf(" [%d]-%s ", panelNum, panel.Title)
	l.renderPanel(v, panel)
	return nil
}

func (l *Layout) renderPanel(v *gocui.View, panel *Panel) {
	v.Clear()
	for i, item := range panel.Items {
		if i == panel.SelectedIndex {
			fmt.Fprintf(v, "> %s\n", item)
		} else {
			fmt.Fprintf(v, "  %s\n", item)
		}
	}
	if len(panel.Items) == 0 {
		fmt.Fprintln(v, "  (empty)")
	}
}

func (l *Layout) createHelpView(g *gocui.Gui, maxX, maxY, helpHeight int) error {
	v, err := g.SetView(panelHelp, 0, maxY-helpHeight-1, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Frame = false
	v.Clear()
	fmt.Fprintln(v, " ←/→: switch panel | ↑/k: up | ↓/j: down | 1-4: jump panel | n: new | q: quit")
	return nil
}

func (l *Layout) getPanelIndex(name string) int {
	for i, p := range l.panels {
		if p.Name == name {
			return i
		}
	}
	return 0
}

func (l *Layout) MoveUp(g *gocui.Gui) {
	panel := l.panels[l.activePanelIndex]
	if panel.SelectedIndex > 0 {
		panel.SelectedIndex--
		l.updateView(g, panel)
	}
}

func (l *Layout) MoveDown(g *gocui.Gui) {
	panel := l.panels[l.activePanelIndex]
	if panel.SelectedIndex < len(panel.Items)-1 {
		panel.SelectedIndex++
		l.updateView(g, panel)
	}
}

func (l *Layout) NextPanel(g *gocui.Gui) {
	l.activePanelIndex = (l.activePanelIndex + 1) % len(l.panels)
	l.refreshAllPanels(g)
}

func (l *Layout) PrevPanel(g *gocui.Gui) {
	l.activePanelIndex--
	if l.activePanelIndex < 0 {
		l.activePanelIndex = len(l.panels) - 1
	}
	l.refreshAllPanels(g)
}

func (l *Layout) JumpToPanel(g *gocui.Gui, index int) {
	if index >= 0 && index < len(l.panels) {
		l.activePanelIndex = index
		l.refreshAllPanels(g)
	}
}

func (l *Layout) updateView(g *gocui.Gui, panel *Panel) {
	g.Update(func(g *gocui.Gui) error {
		if v, err := g.View(panel.Name); err == nil {
			l.renderPanel(v, panel)
		}
		return nil
	})
}

func (l *Layout) refreshAllPanels(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		for i, panel := range l.panels {
			if v, err := g.View(panel.Name); err == nil {
				panelNum := i + 1
				v.Title = fmt.Sprintf(" [%d]-%s ", panelNum, panel.Title)
				l.renderPanel(v, panel)
			}
		}
		activePanel := l.panels[l.activePanelIndex]
		_, _ = g.SetCurrentView(activePanel.Name)
		return nil
	})
}
