package vendingmachine

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializationIsSuccessful(t *testing.T) {
	items := getDefaultItems()

	vm, err := New(items)
	require.NoError(t, err)

	assert.True(t, reflect.DeepEqual(vm.prodmap, getDefaultProdMap()))
}

func TestInsert(t *testing.T) {
	vm, err := New(getDefaultItems())
	require.NoError(t, err)

	amount := 50
	require.NoError(t, vm.InsertCoin(amount))
	assert.Equal(t, Selecting, vm.state)
	assert.Equal(t, amount, *vm.insertedAmount)
	assert.Nil(t, vm.selectedProd)

	// check can not insert in states other than idle
	vm.state = Selecting
	require.ErrorIs(t, vm.InsertCoin(amount), ErrBadState)
	vm.state = Delivering
	require.ErrorIs(t, vm.InsertCoin(amount), ErrBadState)
}

func TestSelectProduct(t *testing.T) {
	amount := 50
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}

	prod := "coffee"
	require.NoError(t, vm.SelectProduct(prod))
	assert.Equal(t, Delivering, vm.state)
	assert.Equal(t, prod, *vm.selectedProd)
	assert.Equal(t, amount, *vm.insertedAmount, "inserted amount should stay the same after selecting")

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	require.ErrorIs(t, vm.SelectProduct("invalid-product"), ErrInvalidProduct)
	assert.Equal(t, Selecting, vm.state)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, amount, *vm.insertedAmount)
	assert.Equal(t, getDefaultProdMap(), vm.prodmap)

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	require.ErrorIs(t, vm.SelectProduct("milk"), ErrOutOfStock)
	assert.Equal(t, Selecting, vm.state)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, amount, *vm.insertedAmount)
	assert.Equal(t, getDefaultProdMap(), vm.prodmap)

	vm = &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Selecting,
		insertedAmount: &amount,
		selectedProd:   nil,
		prodmap:        getDefaultProdMap(),
	}
	require.ErrorIs(t, vm.SelectProduct("coke"), ErrInsufficientFunds)
	assert.Equal(t, Selecting, vm.state)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, amount, *vm.insertedAmount)
	assert.Equal(t, getDefaultProdMap(), vm.prodmap)

	// check can not select in states other than selecting
	vm.state = Idle
	require.ErrorIs(t, vm.SelectProduct("prod"), ErrBadState)
	vm.state = Delivering
	require.ErrorIs(t, vm.SelectProduct("prod"), ErrBadState)
}

func TestDeliverProduct(t *testing.T) {
	amount := 70
	prod := "coffee"
	vm := &VendingMachine{
		mu:             &sync.Mutex{},
		state:          Delivering,
		insertedAmount: &amount,
		selectedProd:   &prod,
		prodmap:        getDefaultProdMap(),
	}

	require.NoError(t, vm.DeliverProduct())
	assert.Equal(t, Idle, vm.state)
	assert.Nil(t, vm.selectedProd)
	assert.Equal(t, 20, *vm.insertedAmount)
	assert.Equal(t, 1, vm.prodmap["coffee"].Number)

	// check can not deliver in states other than delivering
	vm.state = Idle
	require.ErrorIs(t, vm.DeliverProduct(), ErrBadState)
	vm.state = Selecting
	require.ErrorIs(t, vm.DeliverProduct(), ErrBadState)
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
