package domain

import "errors"

var (
	ErrNotFound                     = errors.New("not found")
	ErrAlreadyExists                = errors.New("already exists")
	ErrPaymentEventAlreadyProcessed = errors.New("payment event already processed")
	ErrPaymentAlreadyActivated      = errors.New("payment already activated")
	ErrUnauthorized                 = errors.New("unauthorized")
)
