package pds

import "errors"

var (
	ErrUnavailableMenu = errors.New("Menu unavailable")
	ErrUnknownPizza    = errors.New("Unknown pizza")
	ErrUnknownCustomer = errors.New("Unknown customer")
)
