package viewmodel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

type BrokersViewModel struct {
	mu              sync.RWMutex
	brokers         []models.Broker
	selectedIndex   int
	notifyCh        chan types.ChangeEvent
	commandBindings []*types.CommandBinding
	gui             *gocui.Gui
}

func NewBrokersViewModel(brokers []models.Broker) *BrokersViewModel {
	vm := &BrokersViewModel{
		brokers:       brokers,
		selectedIndex: 0,
		notifyCh:      make(chan types.ChangeEvent),
	}
	vm.initCommandBindings()
	return vm
}

func (vm *BrokersViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *BrokersViewModel) Notify(fieldName string) {
	vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}
}

func (vm *BrokersViewModel) GetSelectedIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedIndex
}

func (vm *BrokersViewModel) SetSelectedIndex(index int) {
	vm.mu.Lock()
	if index >= 0 && index < len(vm.brokers) {
		vm.selectedIndex = index
		vm.mu.Unlock()
		vm.Notify(types.FieldSelectedIndex)
		return
	}
	vm.mu.Unlock()
}

func (vm *BrokersViewModel) GetItemCount() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return len(vm.brokers)
}

func (vm *BrokersViewModel) MoveUp() error {
	vm.mu.Lock()
	if vm.selectedIndex > 0 {
		vm.selectedIndex--
		vm.mu.Unlock()
		vm.Notify(types.FieldSelectedIndex)
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *BrokersViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.brokers)-1 {
		vm.selectedIndex++
		vm.mu.Unlock()
		vm.Notify(types.FieldSelectedIndex)
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *BrokersViewModel) initCommandBindings() {
	moveUp := types.NewCommand(vm.MoveUp)
	moveDown := types.NewCommand(vm.MoveDown)
	openEditor := types.NewCommand(vm.OpenConfigInEditor)

	vm.commandBindings = []*types.CommandBinding{
		{Key: 'k', Cmd: moveUp},
		{Key: 'j', Cmd: moveDown},
		{Key: gocui.KeyArrowUp, Cmd: moveUp},
		{Key: gocui.KeyArrowDown, Cmd: moveDown},
		{Key: 'e', Cmd: openEditor},
	}
}

func (vm *BrokersViewModel) GetCommandBindings() []*types.CommandBinding {
	return vm.commandBindings
}

func (vm *BrokersViewModel) GetDisplayItems() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	items := make([]string, len(vm.brokers))
	for i, b := range vm.brokers {
		items[i] = fmt.Sprintf("%s (%s)", b.Name, b.Address)
	}
	return items
}

func (vm *BrokersViewModel) GetTitle() string {
	return "Brokers"
}

func (vm *BrokersViewModel) GetName() string {
	return "brokers"
}

func (vm *BrokersViewModel) GetSelectedBroker() *models.Broker {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.selectedIndex >= 0 && vm.selectedIndex < len(vm.brokers) {
		return &vm.brokers[vm.selectedIndex]
	}
	return nil
}

func (vm *BrokersViewModel) LoadBrokers(brokers []models.Broker) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.brokers = brokers
	if vm.selectedIndex >= len(brokers) {
		vm.selectedIndex = 0
	}
}

func (vm *BrokersViewModel) AddBrokerConfig(config models.BrokerConfig) {
	vm.mu.Lock()
	newBroker := models.Broker{
		ID:      len(vm.brokers),
		Name:    config.Name,
		Address: config.BootstrapServers,
	}
	vm.brokers = append(vm.brokers, newBroker)
	vm.mu.Unlock()
	vm.Notify(types.FieldItems)
}

func (vm *BrokersViewModel) SetGui(gui *gocui.Gui) {
	vm.gui = gui
}

func (vm *BrokersViewModel) OpenConfigInEditor() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".lazykafka", "brokers.json")

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
