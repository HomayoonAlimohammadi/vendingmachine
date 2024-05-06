package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"

	"vendingmachine/internal/statemachine"
	internalVM "vendingmachine/internal/vendingmachine"
)

type Server struct {
	mu *sync.RWMutex
	// vmMap maps the machine id to the vending machine instance
	vmMap map[string]*internalVM.VendingMachine

	// stMap is the machine id to the statemachine instance
	stMap map[string]*statemachine.Machine
}

func NewAPI() *Server {
	return &Server{
		mu:    &sync.RWMutex{},
		vmMap: make(map[string]*internalVM.VendingMachine),
		stMap: make(map[string]*statemachine.Machine),
	}
}

type AddVMRequest struct {
	Inventory []internalVM.Item `json:"inventory"`
}

type AddVMResponse struct {
	ID string `json:"machine_id"`
}

func (s *Server) AddVMHandler(w http.ResponseWriter, r *http.Request) {
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

	st, err := statemachine.New(req.Inventory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// generate a new machine id
	id := uuid.New().String()
	_, vmFound := s.vmMap[id]
	_, stFound := s.stMap[id]
	isDuplicate := vmFound || stFound
	for isDuplicate {
		id = uuid.New().String()
		_, vmFound = s.vmMap[id]
		_, stFound = s.stMap[id]
		isDuplicate = vmFound || stFound
	}

	s.vmMap[id] = vm
	s.stMap[id] = st

	encode(w, http.StatusOK, AddVMResponse{ID: id})
}

type InsertCoinRequest struct {
	ID     string `json:"machine_id"`
	Amount *int   `json:"inserted_amount"`
}

func (s *Server) InsertCoinHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[InsertCoinRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount == nil {
		http.Error(w, "no amount inserted", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	vm, ok := s.vmMap[req.ID]
	if !ok {
		http.Error(w, fmt.Sprintf("machine id %q not found", req.ID), http.StatusNotFound)
		return
	}
	s.mu.RUnlock()

	err = vm.InsertCoin(*req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("inserted coin successfully"))
}

type SelectProductRequest struct {
	ID      string `json:"machine_id"`
	Product string `json:"selected_product"`
}

func (s *Server) SelectProductHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[SelectProductRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	vm, ok := s.vmMap[req.ID]
	if !ok {
		http.Error(w, fmt.Sprintf("machine id %q not found", req.ID), http.StatusNotFound)
		return
	}
	s.mu.RUnlock()

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

	w.Write([]byte("selected and delivered product successfully"))
}

type AbortOrderRequest struct {
	ID string `json:"machine_id"`
}

func (s *Server) AbortOrderHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[AbortOrderRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	vm, ok := s.vmMap[req.ID]
	if !ok {
		http.Error(w, fmt.Sprintf("machine id %q not found", req.ID), http.StatusNotFound)
		return
	}
	s.mu.RUnlock()

	vm.AbortAndReset()
	w.Write([]byte("aborted successfully"))
}

// #######
// State Machine APIs
// #######

type TransitionRequest struct {
	ID   string            `json:"machine_id"`
	Data statemachine.Data `json:"data"`
}

func (s *Server) TransitionHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decode[TransitionRequest](r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request: ", err.Error()), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	st := s.stMap[req.ID]
	// nil statemachine is no different than not found
	if st == nil {
		http.Error(w, fmt.Sprintf("machine id not found: %q", req.ID), http.StatusNotFound)
		return
	}
	s.mu.RUnlock()

	err = st.Transit(req.Data)
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

	w.WriteHeader(http.StatusOK)
}
