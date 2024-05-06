package statemachine

import "errors"

var (
	ErrInvalidProduct    = errors.New("invalid product")
	ErrOutOfStock        = errors.New("out of stock")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
