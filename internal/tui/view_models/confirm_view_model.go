package viewmodel

import "sync"

type ConfirmViewModel struct {
	mu       sync.RWMutex
	message  string
	onYes    func()
	onNo     func()
	isActive bool
}

func NewConfirmViewModel(message string, onYes, onNo func()) *ConfirmViewModel {
	return &ConfirmViewModel{
		message: message,
		onYes:   onYes,
		onNo:    onNo,
	}
}

func (vm *ConfirmViewModel) GetMessage() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.message
}

func (vm *ConfirmViewModel) SetMessage(msg string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.message = msg
}

func (vm *ConfirmViewModel) Confirm() {
	vm.mu.RLock()
	onYes := vm.onYes
	vm.mu.RUnlock()
	if onYes != nil {
		onYes()
	}
}

func (vm *ConfirmViewModel) Cancel() {
	vm.mu.RLock()
	onNo := vm.onNo
	vm.mu.RUnlock()
	if onNo != nil {
		onNo()
	}
}

func (vm *ConfirmViewModel) GetName() string {
	return "confirm"
}

func (vm *ConfirmViewModel) GetTitle() string {
	return "Confirm"
}
