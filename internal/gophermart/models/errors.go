package models

import "errors"

var (
	ErrLoginPasswordEmpity      = errors.New("login and/or password empity")
	ErrLongLogin                = errors.New("login length is more than 64 characters")
	ErrInvalidOrderNumberFormat = errors.New("invalid order number format")
	ErrWrongWithdrawalSum       = errors.New("withdrawal sum must be greater than zero")
)
