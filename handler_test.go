package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"vendingmachine/internal/statemachine"
	"vendingmachine/internal/storage"
	internalVM "vendingmachine/internal/vendingmachine"
	mock_main "vendingmachine/mocks"
)

func TestAddVMHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vmStorage := getVMStorageMock(t)
		smStorage := getSMStorageMock(t)

		h := NewHandler(vmStorage, smStorage)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/addvm",
			strings.NewReader("{\"inventory\":[{\"name\":\"coke\",\"number\":1,\"price\":100},{\"name\":\"coffee\",\"number\":2,\"price\":50},{\"name\":\"milk\",\"number\":0,\"price\":80}]}"))
		h.AddVMHandler(w, r)
		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, w.Code)
		data, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "{\"machine_id\":\"123\",\"statemachine_id\":\"123\"}\n", string(data))
	})

	t.Run("faulty storage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().SaveVM(gomock.Any()).AnyTimes().Return("", errors.New("some error"))

		smStorage := mock_main.NewMockSMStorage(ctrl)
		smStorage.EXPECT().SaveSM(gomock.Any()).AnyTimes().Return("", errors.New("some error"))

		h := NewHandler(vmStorage, smStorage)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/addvm",
			strings.NewReader("{\"inventory\":[{\"name\":\"coke\",\"number\":1,\"price\":100},{\"name\":\"coffee\",\"number\":2,\"price\":50},{\"name\":\"milk\",\"number\":0,\"price\":80}]}"))
		h.AddVMHandler(w, r)
		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInsertCoinHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vmStorage := getVMStorageMock(t)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/insert",
			strings.NewReader("{\"machine_id\":\"123\", \"inserted_amount\":100}"))
		h.InsertCoinHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no amount", func(t *testing.T) {
		vmStorage := getVMStorageMock(t)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/insert",
			strings.NewReader("{\"machine_id\":\"123\"}"))
		h.InsertCoinHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("faulty storage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(nil, errors.New("some error"))

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/insert",
			strings.NewReader("{\"machine_id\":\"123\", \"inserted_amount\":100}"))

		h.InsertCoinHandler(w, r)
		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(nil, storage.ErrVMNotFound)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/insert",
			strings.NewReader("{\"machine_id\":\"123\", \"inserted_amount\":100}"))

		h.InsertCoinHandler(w, r)
		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSelectProductHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vm := getVMForSelect(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(vm, nil)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/select",
			strings.NewReader("{\"machine_id\":\"123\", \"selected_product\":\"coffee\"}"))
		h.SelectProductHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no product", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vm := getVMForSelect(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(vm, nil)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/select",
			strings.NewReader("{\"machine_id\":\"123\"}"))
		h.SelectProductHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid product", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vm := getVMForSelect(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(vm, nil)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/select",
			strings.NewReader("{\"machine_id\":\"123\", \"selected_product\":\"invalid\"}"))
		h.SelectProductHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("vm not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(nil, storage.ErrVMNotFound)

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/select",
			strings.NewReader("{\"machine_id\":\"123\", \"selected_product\":\"random\"}"))
		h.SelectProductHandler(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("faulty storage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		vmStorage := mock_main.NewMockVMStorage(ctrl)
		vmStorage.EXPECT().GetVM(gomock.Any()).AnyTimes().Return(nil, errors.New("some error"))

		h := NewHandler(vmStorage, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/select",
			strings.NewReader("{\"machine_id\":\"123\", \"selected_product\":\"random\"}"))
		h.SelectProductHandler(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func getVMStorageMock(t *testing.T) *mock_main.MockVMStorage {
	ctrl := gomock.NewController(t)
	m := mock_main.NewMockVMStorage(ctrl)
	m.EXPECT().SaveVM(gomock.Any()).AnyTimes().Return("123", nil)
	vm, err := internalVM.New([]internalVM.Item{
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
	})
	assert.NoError(t, err)
	m.EXPECT().GetVM("123").AnyTimes().Return(vm, nil)

	return m
}

func getSMStorageMock(t *testing.T) *mock_main.MockSMStorage {
	ctrl := gomock.NewController(t)
	m := mock_main.NewMockSMStorage(ctrl)
	sm, err := statemachine.New([]internalVM.Item{
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
	})
	assert.NoError(t, err)
	m.EXPECT().SaveSM(gomock.Any()).AnyTimes().Return("123", nil)
	m.EXPECT().GetSM("123").AnyTimes().Return(sm, nil)

	return m
}

func getVMForSelect(t *testing.T) *internalVM.VendingMachine {
	t.Helper()
	vm, err := internalVM.New(
		nil,
		internalVM.WithState(internalVM.Selecting),
		internalVM.WithInsertedAmount(80),
		internalVM.WithProdMap(
			map[string]*internalVM.Item{
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
			},
		),
	)
	assert.NoError(t, err)

	return vm
}
