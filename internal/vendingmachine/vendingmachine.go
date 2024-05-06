package vendingmachine

import (
	"fmt"
	"sync"
)

type State string

const (
	idle       State = "Idle"
	selecting  State = "Selecting"
	delivering State = "Delivering"
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

func New(inventory []Item) (*VendingMachine, error) {
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          idle,
		insertedAmount: nil,
		prodmap:        make(map[string]*Item),
	}

	// initialize the inventory
	for _, item := range inventory {
		vm.prodmap[item.Name] = &item
	}

	return vm, nil
}

func (vm *VendingMachine) InsertCoin(amount int) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != idle {
		return fmt.Errorf("cannot insert coin in state: %q", vm.state)
	}

	vm.state = selecting
	vm.insertedAmount = &amount

	return nil
}

func (vm *VendingMachine) SelectProduct(productStr string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != selecting {
		return fmt.Errorf("cannot select product in state: %q", vm.state)
	}

	prod, ok := vm.prodmap[productStr]
	if !ok {
		return fmt.Errorf("%w: %q", ErrInvalidProduct, productStr)
	}

	if prod.Number < 1 {
		return fmt.Errorf("%w: product: %q", ErrOutOfStock, prod.Name)
	}

	if vm.insertedAmount != nil && *vm.insertedAmount < prod.Price {
		return fmt.Errorf("%w: product: %q, price: %d, inserted amount: %d", ErrInsufficientFunds, prod.Name, prod.Price, *vm.insertedAmount)
	}

	vm.state = delivering
	vm.selectedProd = &productStr

	return nil
}

func (vm *VendingMachine) DeliverProduct() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.state != delivering {
		return fmt.Errorf("cannot deliver product in state: %q", vm.state)
	}

	if vm.selectedProd == nil {
		return fmt.Errorf("no product was selected")
	}

	// reset
	vm.state = idle
	vm.insertedAmount = nil
	// reduce the number of product in the inventory
	prod := vm.prodmap[*vm.selectedProd]
	if prod == nil {
		return fmt.Errorf("no product to deliver")
	}
	prod.Number -= 1
	vm.selectedProd = nil

	return nil
}

func (vm *VendingMachine) AbortAndReset() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.state = idle
	vm.insertedAmount = nil
	vm.selectedProd = nil
}
