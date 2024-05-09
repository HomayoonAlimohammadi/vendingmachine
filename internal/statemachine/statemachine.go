package statemachine

import (
	"errors"
	"fmt"
	"sync"

	"vendingmachine/internal/vendingmachine"
)

type Data struct {
	// InsertedAmount specifies the amount of coins inserted in the
	// previous steps, nil means no coins
	InsertedAmount *int `json:"inserted_amount"`

	// SelectedProd indicates the product selected in the previous step
	// nil means the previous step was not a product selection
	SelectedProd *string `json:"selected_product"`

	// SelectedProdProb specifies the properties of the selected product
	// nil means no product is selected
	selectedProdProb *vendingmachine.Item

	// product to properties map
	prodMap map[string]*vendingmachine.Item
}

type Machine struct {
	currentState State
	mu           *sync.Mutex
	data         *Data
}

func New(items []vendingmachine.Item) (*Machine, error) {
	cs := &idleState{}
	m := &Machine{
		mu:           &sync.Mutex{},
		currentState: cs,
		data: &Data{
			prodMap: make(map[string]*vendingmachine.Item),
		},
	}

	for _, item := range items {
		m.data.prodMap[item.Name] = &item
	}

	cs.m = m

	err := cs.m.currentState.Enter()
	if err != nil {
		return nil, fmt.Errorf("failed to enter initial state: %w", err)
	}

	return m, nil
}

func (m *Machine) Transit(d Data) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.currentState.Transit(m, d)
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

func (m *Machine) setState(s State) error {
	m.currentState = s

	err := m.currentState.Enter()
	if err != nil {
		return fmt.Errorf("failed to enter state: %w", err)
	}

	return nil
}

type State interface {
	Transit(*Machine, Data) error
	Enter() error
}

type idleState struct {
	m *Machine
}

func (s *idleState) Enter() error {
	// slog.Info("entered idle state")
	// b, _ := json.MarshalIndent(s.m.data, "", "  ")
	// slog.Info(string(b))
	return nil
}

func (s *idleState) Transit(m *Machine, d Data) error {
	if d.InsertedAmount == nil {
		return errors.New("no coins inserted")
	}

	s.m.data.InsertedAmount = d.InsertedAmount

	err := m.setState(&selectingState{m: m})
	if err != nil {
		return fmt.Errorf("failed to set state to 'selecting': %w", err)
	}

	return nil
}

type selectingState struct {
	m *Machine
}

func (s *selectingState) Enter() error {
	// slog.Info("entered selecing state")
	// b, _ := json.MarshalIndent(s.m.data, "", "  ")
	// slog.Info(string(b))
	return nil
}

func (s *selectingState) Transit(m *Machine, d Data) error {
	if d.SelectedProd == nil {
		return errors.New("no product was selected")
	}

	prop := s.m.data.prodMap[*d.SelectedProd]
	if prop == nil {
		return fmt.Errorf("%w: %q", ErrInvalidProduct, *d.SelectedProd)
	}

	if prop.Number < 1 {
		return fmt.Errorf("%w: product: %q", ErrOutOfStock, *d.SelectedProd)
	}

	if s.m.data.InsertedAmount == nil || *s.m.data.InsertedAmount < prop.Price {
		return fmt.Errorf("%w: product: %q, price: %d, inserted amount: %d",
			ErrInsufficientFunds, *d.SelectedProd, prop.Price, *s.m.data.InsertedAmount)
	}

	s.m.data.SelectedProd = d.SelectedProd
	s.m.data.selectedProdProb = prop

	err := m.setState(&deliveringState{m: m})
	if err != nil {
		return fmt.Errorf("failed to set state to 'delivering': %w", err)
	}

	return nil
}

type deliveringState struct {
	m *Machine
}

func (s *deliveringState) Enter() error {
	// slog.Info("entered delivering state")
	// b, _ := json.MarshalIndent(s.m.data, "", "  ")
	// slog.Info(string(b))
	return nil
}

func (s *deliveringState) Transit(m *Machine, _ Data) error {
	if s.m.data.selectedProdProb == nil {
		return errors.New("selected product not found")
	}

	s.m.data.selectedProdProb.Number--

	// reset state
	s.m.data.InsertedAmount = nil
	s.m.data.SelectedProd = nil
	s.m.data.selectedProdProb = nil

	err := m.setState(&idleState{m: m})
	if err != nil {
		return fmt.Errorf("failed to set state to 'idle': %w", err)
	}

	return nil
}
