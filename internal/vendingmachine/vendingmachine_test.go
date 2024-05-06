package vendingmachine

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializationIsSuccessful(t *testing.T) {
	items := getDefaultItems()

	vm, err := New(items)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(vm.prodmap, getDefaultProdMap()))
}

func TestInsert(t *testing.T) {
	vm, err := New(getDefaultItems())
	assert.NoError(t, err)

	amount := 50
	assert.NoError(t, vm.InsertCoin(amount))
	assert.Equal(t, vm.state, selecting)
	assert.Equal(t, *vm.insertedAmount, amount)
	assert.Nil(t, vm.selectedProd)

	// check can not insert in states other than idle
	vm.state = selecting
	assert.ErrorIs(t, vm.InsertCoin(amount), ErrBadState)
	vm.state = delivering
	assert.ErrorIs(t, vm.InsertCoin(amount), ErrBadState)
}

func TestSelectProduct(t *testing.T) {
	amount := 50
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}

	prod := "coffee"
	assert.NoError(t, vm.SelectProduct(prod))
	assert.Equal(t, vm.state, delivering)
	assert.Equal(t, vm.selectedProd, &prod)
	assert.Equal(t, *vm.insertedAmount, amount, "inserted amount should stay the same after selecting")

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	assert.ErrorIs(t, vm.SelectProduct("invalid-product"), ErrInvalidProduct)
	assert.Equal(t, vm.state, selecting)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, *vm.insertedAmount, amount)
	assert.Equal(t, vm.prodmap, getDefaultProdMap())

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	assert.ErrorIs(t, vm.SelectProduct("milk"), ErrOutOfStock)
	assert.Equal(t, vm.state, selecting)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, *vm.insertedAmount, amount)
	assert.Equal(t, vm.prodmap, getDefaultProdMap())

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	assert.ErrorIs(t, vm.SelectProduct("coke"), ErrInsufficientFunds)
	assert.Equal(t, vm.state, selecting)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, *vm.insertedAmount, amount)
	assert.Equal(t, vm.prodmap, getDefaultProdMap())

	// check can not select in states other than selecting
	vm.state = idle
	assert.ErrorIs(t, vm.SelectProduct("prod"), ErrBadState)
	vm.state = delivering
	assert.ErrorIs(t, vm.SelectProduct("prod"), ErrBadState)
}

func TestDeliverProduct(t *testing.T) {
	amount := 70
	prod := "coffee"
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          delivering,
		insertedAmount: &amount,
		selectedProd:   &prod,
		prodmap:        getDefaultProdMap(),
	}

	assert.NoError(t, vm.DeliverProduct())
	assert.Equal(t, vm.state, idle)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, *vm.insertedAmount, 20)
	assert.Equal(t, vm.prodmap["coffee"].Number, 1)

	// check can not deliver in states other than delivering
	vm.state = idle
	assert.ErrorIs(t, vm.DeliverProduct(), ErrBadState)
	vm.state = selecting
	assert.ErrorIs(t, vm.DeliverProduct(), ErrBadState)
}

func getDefaultItems() []Item {
	return []Item{
		{
			Name:   "coke",
			Number: 1,
			Price:  100,
		},
		{
			Name:   "coffee",
			Number: 2,
			Price:  50,
		},
		{
			Name:   "milk",
			Number: 0,
			Price:  80,
		},
	}
}

func getDefaultProdMap() map[string]*Item {
	prodmap := map[string]*Item{
		"coke": {
			Name:   "coke",
			Number: 1,
			Price:  100,
		},
		"coffee": {
			Name:   "coffee",
			Number: 2,
			Price:  50,
		},
		"milk": {
			Name:   "milk",
			Number: 0,
			Price:  80,
		},
	}

	return prodmap
}
