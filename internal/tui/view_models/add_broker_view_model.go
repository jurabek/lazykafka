package viewmodel

import (
	"errors"
	"regexp"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

var bootstrapServersPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*:\d{1,5}(,[a-zA-Z0-9][a-zA-Z0-9.-]*:\d{1,5})*$`)

const (
	StepName             = 0
	StepBootstrapServers = 1
	StepAuthType         = 2
	StepUsername         = 3
	StepPassword         = 4
)

var ErrValidation = errors.New("validation error")

type AddBrokerViewModel struct {
	mu               sync.RWMutex
	name             string
	bootstrapServers string
	authType         models.AuthType
	username         string
	password         string
	currentStep      int
	notifyCh         chan types.ChangeEvent
	onSubmit         func(config models.BrokerConfig)
	onCancel         func()
}

func NewAddBrokerViewModel(onSubmit func(models.BrokerConfig), onCancel func()) *AddBrokerViewModel {
	return &AddBrokerViewModel{
		currentStep: StepName,
		authType:    models.AuthNone,
		notifyCh:    make(chan types.ChangeEvent, 1),
		onSubmit:    onSubmit,
		onCancel:    onCancel,
	}
}

func (vm *AddBrokerViewModel) NotifyChannel() <-chan types.ChangeEvent {
	return vm.notifyCh
}

func (vm *AddBrokerViewModel) Notify(fieldName string) {
	select {
	case vm.notifyCh <- types.ChangeEvent{FieldName: fieldName}:
	default:
	}
}

func (vm *AddBrokerViewModel) GetCurrentStep() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentStep
}

func (vm *AddBrokerViewModel) GetStepTitle() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	switch vm.currentStep {
	case StepName:
		return "Broker name:"
	case StepBootstrapServers:
		return "Bootstrap servers:"
	case StepAuthType:
		return "Auth type (↑↓ to select, Enter to confirm):"
	case StepUsername:
		return "Username:"
	case StepPassword:
		return "Password:"
	}
	return ""
}

func (vm *AddBrokerViewModel) NextStep() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	switch vm.currentStep {
	case StepName:
		vm.currentStep = StepBootstrapServers
	case StepBootstrapServers:
		vm.currentStep = StepAuthType
	case StepAuthType:
		if vm.authType == models.AuthSASL || vm.authType == models.AuthAWSIAM {
			vm.currentStep = StepUsername
		} else {
			return true // done, submit
		}
	case StepUsername:
		vm.currentStep = StepPassword
	case StepPassword:
		return true // done, submit
	}
	return false
}

func (vm *AddBrokerViewModel) PrevStep() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	switch vm.currentStep {
	case StepBootstrapServers:
		vm.currentStep = StepName
	case StepAuthType:
		vm.currentStep = StepBootstrapServers
	case StepUsername:
		vm.currentStep = StepAuthType
	case StepPassword:
		vm.currentStep = StepUsername
	}
}

func (vm *AddBrokerViewModel) GetAuthTypeOptions() []string {
	return []string{"None", "SASL", "SSL", "AWS IAM"}
}

func (vm *AddBrokerViewModel) GetSelectedAuthTypeIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return int(vm.authType)
}

func (vm *AddBrokerViewModel) SetSelectedAuthTypeIndex(index int) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if index >= 0 && index < 4 {
		vm.authType = models.AuthType(index)
		vm.Notify("authType")
	}
}

func (vm *AddBrokerViewModel) MoveAuthTypeUp() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.authType > 0 {
		vm.authType--
	} else {
		vm.authType = 3 // Wrap to AWS IAM
	}
	vm.Notify("authType")
}

func (vm *AddBrokerViewModel) MoveAuthTypeDown() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.authType < 3 {
		vm.authType++
	} else {
		vm.authType = 0 // Wrap to None
	}
	vm.Notify("authType")
}

func (vm *AddBrokerViewModel) CycleAuthType() {
	vm.MoveAuthTypeDown()
}

func (vm *AddBrokerViewModel) GetAuthType() models.AuthType {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.authType
}

func (vm *AddBrokerViewModel) SetName(name string) {
	vm.mu.Lock()
	vm.name = name
	vm.mu.Unlock()
}

func (vm *AddBrokerViewModel) GetName() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.name
}

func (vm *AddBrokerViewModel) SetBootstrapServers(servers string) {
	vm.mu.Lock()
	vm.bootstrapServers = servers
	vm.mu.Unlock()
}

func (vm *AddBrokerViewModel) GetBootstrapServers() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.bootstrapServers
}

func (vm *AddBrokerViewModel) SetUsername(username string) {
	vm.mu.Lock()
	vm.username = username
	vm.mu.Unlock()
}

func (vm *AddBrokerViewModel) GetUsername() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.username
}

func (vm *AddBrokerViewModel) SetPassword(password string) {
	vm.mu.Lock()
	vm.password = password
	vm.mu.Unlock()
}

func (vm *AddBrokerViewModel) GetPassword() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.password
}

func (vm *AddBrokerViewModel) Validate() error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if strings.TrimSpace(vm.name) == "" {
		return errors.Join(ErrValidation, errors.New("name is required"))
	}
	servers := strings.TrimSpace(vm.bootstrapServers)
	if servers == "" {
		return errors.Join(ErrValidation, errors.New("bootstrap servers is required"))
	}
	if !bootstrapServersPattern.MatchString(servers) {
		return errors.Join(ErrValidation, errors.New("invalid format, use host:port or host:port,host:port"))
	}
	if vm.authType == models.AuthSASL {
		if strings.TrimSpace(vm.username) == "" {
			return errors.Join(ErrValidation, errors.New("username is required for SASL"))
		}
		if strings.TrimSpace(vm.password) == "" {
			return errors.Join(ErrValidation, errors.New("password is required for SASL"))
		}
	}
	if vm.authType == models.AuthAWSIAM {
		if strings.TrimSpace(vm.username) == "" {
			return errors.Join(ErrValidation, errors.New("username is required for AWS IAM"))
		}
		if strings.TrimSpace(vm.password) == "" {
			return errors.Join(ErrValidation, errors.New("password is required for AWS IAM"))
		}
	}
	return nil
}

func (vm *AddBrokerViewModel) Submit() error {
	if err := vm.Validate(); err != nil {
		return err
	}

	vm.mu.RLock()
	config := models.BrokerConfig{
		Name:             strings.TrimSpace(vm.name),
		BootstrapServers: strings.TrimSpace(vm.bootstrapServers),
		AuthType:         vm.authType,
		Username:         strings.TrimSpace(vm.username),
		Password:         strings.TrimSpace(vm.password),
	}
	vm.mu.RUnlock()

	if vm.onSubmit != nil {
		vm.onSubmit(config)
	}
	return nil
}

func (vm *AddBrokerViewModel) Cancel() {
	if vm.onCancel != nil {
		vm.onCancel()
	}
}
