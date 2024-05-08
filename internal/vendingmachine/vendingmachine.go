package vendingmachine

import (
	"fmt"
	"sync"
)

type State string

const (
	// Ready to insert coins
	Idle State = "Idle"

	// Ready to select product
	Selecting State = "Selecting"

	// Ready to deliver the selected producdt
	Delivering State = "Delivering"
)

type VendingMachine struct {
	mu    *sync.Mutex
	state State

	// insertedAmount specifies the amount of coins inserted in the
	// previous steps, nil means no coins
	insertedAmount *int

	// selectedProd indicates the product selected in the previous step
	// nil means the previous step was not a product selection
	selectedProd *string

	// product to properties map
	prodmap map[string]*Item
}

type Item struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
	Price  int    `json:"price"`
}

// VMOption is used to initialize the VendingMachine instance with custom data
// should be used for testing purposes only
type VMOption interface {
	apply(*VendingMachine)
}

type vmOption func(*VendingMachine)

func (o vmOption) apply(vm *VendingMachine) {
	o(vm)
}

func WithState(s State) VMOption {
	return vmOption(func(vm *VendingMachine) {
		vm.state = s
	})
}

func WithInsertedAmount(a int) VMOption {
	return vmOption(func(vm *VendingMachine) {
		vm.insertedAmount = &a
	})
}

func WithSelectedProd(p string) VMOption {
	return vmOption(func(vm *VendingMachine) {
		vm.selectedProd = &p
	})
}

func WithProdMap(m map[string]*Item) VMOption {
	return vmOption(func(vm *VendingMachine) {
		vm.prodmap = m
	})
}

func New(inventory []Item, opts ...VMOption) (*VendingMachine, error) {
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Idle,
		insertedAmount: nil,
		prodmap:        make(map[string]*Item),
	}

	// initialize the inventory
	for _, item := range inventory {
		vm.prodmap[item.Name] = &item
	}

	// overwrite from the options
	for _, o := range opts {
		o.apply(vm)
	}

	return vm, nil
}

func (vm *VendingMachine) InsertCoin(amount int) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != Idle {
		return fmt.Errorf("%w: cannot insert coin in state: %q", ErrBadState, vm.state)
	}

	vm.state = Selecting
	vm.insertedAmount = &amount

	return nil
}

func (vm *VendingMachine) SelectProduct(productStr string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != Selecting {
		return fmt.Errorf("%w: cannot select product in state: %q", ErrBadState, vm.state)
	}

	prod, ok := vm.prodmap[productStr]
	if !ok {
		return fmt.Errorf("%w: %q", ErrInvalidProduct, productStr)
	}

	if prod.Number < 1 {
		return fmt.Errorf("%w: product: %q", ErrOutOfStock, prod.Name)
	}

	// should not happen but check for it anyways
	if vm.insertedAmount == nil {
		return fmt.Errorf("no money was inserted")
	}

	if *vm.insertedAmount < prod.Price {
		return fmt.Errorf("%w: product: %q, price: %d, inserted amount: %d", ErrInsufficientFunds, prod.Name, prod.Price, *vm.insertedAmount)
	}

	vm.state = Delivering
	vm.selectedProd = &productStr

	return nil
}

func (vm *VendingMachine) DeliverProduct() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != Delivering {
		return fmt.Errorf("%w: cannot deliver product in state: %q", ErrBadState, vm.state)
	}

	// should not happen but check for it anyways
	if vm.selectedProd == nil {
		return fmt.Errorf("no product was selected")
	}

	// should not happen but check for it anyways
	if vm.insertedAmount == nil {
		return fmt.Errorf("no money was inserted")
	}

	prod := vm.prodmap[*vm.selectedProd]
	// should not happen but check for it anyways
	if prod == nil {
		return fmt.Errorf("no product to deliver")
	}

	// should not happen but check for it anyways
	if *vm.insertedAmount < prod.Price {
		return fmt.Errorf("not enough money")
	}

	// reset
	vm.state = Idle
	// reduce the number of product in the inventory
	prod.Number -= 1
	*vm.insertedAmount -= prod.Price
	vm.selectedProd = nil

	return nil
}

func (vm *VendingMachine) AbortAndReset() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.state = Idle
	vm.insertedAmount = nil
	vm.selectedProd = nil
}
