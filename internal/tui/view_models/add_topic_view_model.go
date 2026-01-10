package viewmodel

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/jurabek/lazykafka/internal/models"
	"github.com/jurabek/lazykafka/internal/tui/types"
)

const (
	StepTopicName         = 0
	StepPartitions        = 1
	StepReplicationFactor = 2
	StepCleanupPolicy     = 3
	StepMinISR            = 4
	StepRetention         = 5
)

type AddTopicViewModel struct {
	mu                sync.RWMutex
	name              string
	partitions        string
	replicationFactor string
	cleanupPolicy     models.CleanupPolicy
	minISR            string
	retention         string
	currentStep       int
	onChange          types.OnChangeFunc
	onSubmit          func(config models.TopicConfig)
	onCancel          func()
}

func NewAddTopicViewModel(onSubmit func(models.TopicConfig), onCancel func()) *AddTopicViewModel {
	return &AddTopicViewModel{
		currentStep:   StepTopicName,
		cleanupPolicy: models.CleanupDelete,
		minISR:        "1",
		onSubmit:      onSubmit,
		onCancel:      onCancel,
	}
}

func (vm *AddTopicViewModel) SetOnChange(fn types.OnChangeFunc) {
	vm.onChange = fn
}

func (vm *AddTopicViewModel) notifyChange(fieldName string) {
	if vm.onChange != nil {
		vm.onChange(types.ChangeEvent{FieldName: fieldName})
	}
}

func (vm *AddTopicViewModel) GetCurrentStep() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.currentStep
}

func (vm *AddTopicViewModel) GetStepTitle() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	switch vm.currentStep {
	case StepTopicName:
		return "Topic name:"
	case StepPartitions:
		return "Number of partitions:"
	case StepReplicationFactor:
		return "Replication factor:"
	case StepCleanupPolicy:
		return "Cleanup policy (use arrow keys, Enter to confirm):"
	case StepMinISR:
		return "Min in-sync replicas:"
	case StepRetention:
		return "Retention time (e.g., 7d, 168h, or leave empty):"
	}
	return ""
}

func (vm *AddTopicViewModel) NextStep() bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.currentStep >= StepRetention {
		return true
	}
	vm.currentStep++
	return false
}

func (vm *AddTopicViewModel) PrevStep() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.currentStep > StepTopicName {
		vm.currentStep--
	}
}

func (vm *AddTopicViewModel) GetCleanupPolicyOptions() []string {
	return []string{"Delete", "Compact", "Compact-Delete"}
}

func (vm *AddTopicViewModel) GetSelectedCleanupPolicyIndex() int {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return int(vm.cleanupPolicy)
}

func (vm *AddTopicViewModel) MoveCleanupPolicyUp() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.cleanupPolicy > 0 {
		vm.cleanupPolicy--
	} else {
		vm.cleanupPolicy = 2
	}
	vm.notifyChange("cleanupPolicy")
}

func (vm *AddTopicViewModel) MoveCleanupPolicyDown() {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if vm.cleanupPolicy < 2 {
		vm.cleanupPolicy++
	} else {
		vm.cleanupPolicy = 0
	}
	vm.notifyChange("cleanupPolicy")
}

func (vm *AddTopicViewModel) SetName(name string) {
	vm.mu.Lock()
	vm.name = name
	vm.mu.Unlock()
}

func (vm *AddTopicViewModel) GetName() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.name
}

func (vm *AddTopicViewModel) SetPartitions(p string) {
	vm.mu.Lock()
	vm.partitions = p
	vm.mu.Unlock()
}

func (vm *AddTopicViewModel) GetPartitions() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.partitions
}

func (vm *AddTopicViewModel) SetReplicationFactor(rf string) {
	vm.mu.Lock()
	vm.replicationFactor = rf
	vm.mu.Unlock()
}

func (vm *AddTopicViewModel) GetReplicationFactor() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.replicationFactor
}

func (vm *AddTopicViewModel) SetMinISR(isr string) {
	vm.mu.Lock()
	vm.minISR = isr
	vm.mu.Unlock()
}

func (vm *AddTopicViewModel) GetMinISR() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.minISR
}

func (vm *AddTopicViewModel) SetRetention(r string) {
	vm.mu.Lock()
	vm.retention = r
	vm.mu.Unlock()
}

func (vm *AddTopicViewModel) GetRetention() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.retention
}

func (vm *AddTopicViewModel) Validate() error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if strings.TrimSpace(vm.name) == "" {
		return errors.Join(ErrValidation, errors.New("topic name is required"))
	}

	partitions, err := strconv.Atoi(strings.TrimSpace(vm.partitions))
	if err != nil || partitions <= 0 {
		return errors.Join(ErrValidation, errors.New("partitions must be a positive integer"))
	}

	rf, err := strconv.Atoi(strings.TrimSpace(vm.replicationFactor))
	if err != nil || rf <= 0 {
		return errors.Join(ErrValidation, errors.New("replication factor must be a positive integer"))
	}

	isr, err := strconv.Atoi(strings.TrimSpace(vm.minISR))
	if err != nil || isr <= 0 {
		return errors.Join(ErrValidation, errors.New("min ISR must be a positive integer"))
	}

	if isr > rf {
		return errors.Join(ErrValidation, errors.New("min ISR cannot exceed replication factor"))
	}

	retention := strings.TrimSpace(vm.retention)
	if retention != "" {
		if _, err := models.ParseRetention(retention); err != nil {
			return errors.Join(ErrValidation, errors.New("invalid retention format"))
		}
	}

	return nil
}

func (vm *AddTopicViewModel) Submit() error {
	if err := vm.Validate(); err != nil {
		return err
	}

	vm.mu.RLock()
	partitions, err := strconv.Atoi(strings.TrimSpace(vm.partitions))
	if err != nil {
		// if parsing fails, default to 1
		partitions = 1
	}
	rf, err := strconv.Atoi(strings.TrimSpace(vm.replicationFactor))
	if err != nil {
		// if parsing fails, default to 1
		rf = 1
	}

	isr, err := strconv.Atoi(strings.TrimSpace(vm.minISR))
	if err != nil {
		// if parsing fails in case of empty string, default to 1
		isr = 1
	}

	retentionMs, err := models.ParseRetention(strings.TrimSpace(vm.retention))
	if err != nil {
		// if parsing fails, default to 0
		retentionMs = 7 * models.MillsPerDay
	}

	config := models.TopicConfig{
		Name:              strings.TrimSpace(vm.name),
		Partitions:        partitions,
		ReplicationFactor: rf,
		CleanupPolicy:     vm.cleanupPolicy,
		MinInSyncReplicas: isr,
		RetentionMs:       retentionMs,
	}
	vm.mu.RUnlock()

	if vm.onSubmit != nil {
		vm.onSubmit(config)
	}
	return nil
}

func (vm *AddTopicViewModel) Cancel() {
	if vm.onCancel != nil {
		vm.onCancel()
	}
}
