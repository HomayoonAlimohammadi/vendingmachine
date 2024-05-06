package vendingmachine

import "errors"

var (
	ErrBadState          = errors.New("bad state")
	ErrInvalidProduct    = errors.New("invalid product")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrOutOfStock        = errors.New("out of stock")
)
