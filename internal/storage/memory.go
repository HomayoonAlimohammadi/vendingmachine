package storage

import (
	"sync"

	"github.com/google/uuid"

	"vendingmachine/internal/statemachine"
	internalVM "vendingmachine/internal/vendingmachine"
)

type InMemoryVMStorage struct {
	mu sync.RWMutex
	// vmMap maps the machine id to the vending machine instance
	vmMap map[string]*internalVM.VendingMachine
}

func NewInMemoryVMStorage() *InMemoryVMStorage {
	return &InMemoryVMStorage{
		mu:    sync.RWMutex{},
		vmMap: make(map[string]*internalVM.VendingMachine),
	}
}

func (s *InMemoryVMStorage) GetVM(id string) (*internalVM.VendingMachine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vm, ok := s.vmMap[id]
	if !ok {
		return nil, ErrVMNotFound
	}

	return vm, nil
}

func (s *InMemoryVMStorage) SaveVM(vm *internalVM.VendingMachine) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// generate a new machine id
	id := uuid.New().String()
	_, isDuplicate := s.vmMap[id]
	for isDuplicate {
		id = uuid.New().String()
		_, isDuplicate = s.vmMap[id]
	}

	s.vmMap[id] = vm

	return id, nil
}

type InMemorySMStorage struct {
	mu sync.RWMutex
	// stMap is the machine id to the statemachine instance
	smMap map[string]*statemachine.Machine
}

func NewInMemorySMStorage() *InMemorySMStorage {
	return &InMemorySMStorage{
		mu:    sync.RWMutex{},
		smMap: make(map[string]*statemachine.Machine),
	}
}

func (s *InMemorySMStorage) GetSM(id string) (*statemachine.Machine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sm, ok := s.smMap[id]
	if !ok {
		return nil, ErrVMNotFound
	}

	return sm, nil
}

func (s *InMemorySMStorage) SaveSM(sm *statemachine.Machine) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// generate a new machine id
	id := uuid.New().String()
	_, isDuplicate := s.smMap[id]
	for isDuplicate {
		id = uuid.New().String()
		_, isDuplicate = s.smMap[id]
	}

	s.smMap[id] = sm

	return id, nil
}
