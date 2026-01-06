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

type BrokerSelectionChangedFunc func(broker *models.Broker)

type BrokersViewModel struct {
	mu                 sync.RWMutex
	brokers            []models.Broker
	selectedIndex      int
	onChange           types.OnChangeFunc
	commandBindings    []*types.CommandBinding
	gui                *gocui.Gui
	onSelectionChanged BrokerSelectionChangedFunc
}

func NewBrokersViewModel() *BrokersViewModel {
	vm := &BrokersViewModel{
		selectedIndex: -1,
	}
	vm.initCommandBindings()
	return vm
}

func (vm *BrokersViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *BrokersViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
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
		broker := &vm.brokers[index]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		vm.notifyChange(types.FieldSelectedIndex)
		if callback != nil {
			callback(broker)
		}
		return
	}
	vm.mu.Unlock()
}

func (vm *BrokersViewModel) SetOnSelectionChanged(fn BrokerSelectionChangedFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.onSelectionChanged = fn
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
		broker := &vm.brokers[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		vm.notifyChange(types.FieldSelectedIndex)
		if callback != nil {
			callback(broker)
		}
		return nil
	}
	vm.mu.Unlock()
	return types.ErrNoSelection
}

func (vm *BrokersViewModel) MoveDown() error {
	vm.mu.Lock()
	if vm.selectedIndex < len(vm.brokers)-1 {
		vm.selectedIndex++
		broker := &vm.brokers[vm.selectedIndex]
		callback := vm.onSelectionChanged
		vm.mu.Unlock()
		vm.notifyChange(types.FieldSelectedIndex)
		if callback != nil {
			callback(broker)
		}
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

func (vm *BrokersViewModel) Load(brokers []models.Broker) {
	vm.mu.Lock()
	vm.brokers = brokers
	vm.selectedIndex = -1
	vm.mu.Unlock()

	vm.notifyChange(types.FieldItems)
	vm.SetSelectedIndex(0)
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
	vm.notifyChange(types.FieldItems)
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
