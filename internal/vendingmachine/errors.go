package vendingmachine

import "errors"

var (
	ErrInvalidProduct    = errors.New("invalid product")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrOutOfStock        = errors.New("out of stock")
)
