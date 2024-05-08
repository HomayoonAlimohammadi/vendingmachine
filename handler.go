package main

import (
	"errors"
	"fmt"
	"net/http"

	"vendingmachine/internal/statemachine"
	"vendingmachine/internal/storage"
	internalVM "vendingmachine/internal/vendingmachine"
)

type VMStorage interface {
	GetVM(id string) (*internalVM.VendingMachine, error)
	SaveVM(vm *internalVM.VendingMachine) (id string, err error)
}

type SMStorage interface {
	GetSM(id string) (*statemachine.Machine, error)
	SaveSM(sm *statemachine.Machine) (id string, err error)
}

type Handler struct {
	vmStorage VMStorage
	smStorage SMStorage
}

func NewHandler(vmStorage VMStorage, smStorage SMStorage) *Handler {
	return &Handler{
		vmStorage: vmStorage,
		smStorage: smStorage,
	}
}

type AddVMRequest struct {
	Inventory []internalVM.Item `json:"inventory"`
}

type AddVMResponse struct {
	VMID string `json:"machine_id"`
	SMID string `json:"statemachine_id"`
}

func (s *Handler) AddVMHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[AddVMRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vm, err := internalVM.New(req.Inventory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vmID, err := s.vmStorage.SaveVM(vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sm, err := statemachine.New(req.Inventory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	smID, err := s.smStorage.SaveSM(sm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encode(w, http.StatusOK, AddVMResponse{VMID: vmID, SMID: smID})
}

type InsertCoinRequest struct {
	ID     string `json:"machine_id"`
	Amount int    `json:"inserted_amount"`
}

func (s *Handler) InsertCoinHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[InsertCoinRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount == 0 {
		http.Error(w, "no amount inserted", http.StatusBadRequest)
		return
	}

	vm, err := s.vmStorage.GetVM(req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrVMNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = vm.InsertCoin(req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("inserted coin successfully"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type SelectProductRequest struct {
	ID      string `json:"machine_id"`
	Product string `json:"selected_product"`
}

func (s *Handler) SelectProductHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[SelectProductRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Product == "" {
		http.Error(w, "no product was selected", http.StatusBadRequest)
		return
	}

	vm, err := s.vmStorage.GetVM(req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrVMNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = vm.SelectProduct(req.Product)
	if err != nil {
		if errors.Is(err, internalVM.ErrInvalidProduct) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, internalVM.ErrInsufficientFunds) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = vm.DeliverProduct()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("selected and delivered product successfully"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type AbortOrderRequest struct {
	ID string `json:"machine_id"`
}

func (s *Handler) AbortOrderHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[AbortOrderRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vm, err := s.vmStorage.GetVM(req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrVMNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm.AbortAndReset()
	_, err = w.Write([]byte("aborted successfully"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// #######
// State Machine Handlers
// #######

type TransitionRequest struct {
	ID   string            `json:"machine_id"`
	Data statemachine.Data `json:"data"`
}

func (s *Handler) TransitionHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[TransitionRequest](r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request: %+v", err.Error()), http.StatusBadRequest)
		return
	}

	sm, err := s.smStorage.GetSM(req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrSMNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = sm.Transit(req.Data)
	if err != nil {
		if errors.Is(err, statemachine.ErrInvalidProduct) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, statemachine.ErrOutOfStock) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, statemachine.ErrInsufficientFunds) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("transitioned successfully"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
