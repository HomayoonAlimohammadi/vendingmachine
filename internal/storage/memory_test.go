package storage_test

import (
	"testing"

	"vendingmachine/internal/statemachine"
	"vendingmachine/internal/storage"
	internalVM "vendingmachine/internal/vendingmachine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryVMStorage(t *testing.T) {
	s := storage.NewInMemoryVMStorage()
	vm, err := internalVM.New(getDefaultItems())
	require.NoError(t, err)
	assert.NotNil(t, vm)
	id, err := s.SaveVM(vm)
	require.NoError(t, err)

	fetchedVM, err := s.GetVM(id)
	require.NoError(t, err)
	assert.NotNil(t, fetchedVM)
}

func TestInMemorySMStorage(t *testing.T) {
	s := storage.NewInMemorySMStorage()
	sm, err := statemachine.New(getDefaultItems())
	require.NoError(t, err)
	assert.NotNil(t, sm)
	id, err := s.SaveSM(sm)
	require.NoError(t, err)

	fetchedSM, err := s.GetSM(id)
	require.NoError(t, err)
	assert.NotNil(t, fetchedSM)
}

func getDefaultItems() []internalVM.Item {
	return []internalVM.Item{
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
