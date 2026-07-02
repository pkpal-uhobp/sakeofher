package domain

import "errors"

var (
	ErrNotFound                     = errors.New("not found")
	ErrAlreadyExists                = errors.New("already exists")
	ErrPaymentEventAlreadyProcessed = errors.New("payment event already processed")
	ErrPaymentAlreadyActivated      = errors.New("payment already activated")
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrInvalidInput                 = errors.New("invalid input")
	ErrInactiveTariffPrice          = errors.New("tariff price is inactive")
	ErrPaymentNotPaid               = errors.New("payment is not paid")
)
