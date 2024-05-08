package storage

import "errors"

var (
	ErrVMNotFound = errors.New("vending machine not found")
	ErrSMNotFound = errors.New("state machine not found")
)
