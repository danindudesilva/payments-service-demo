package domain

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid payment status transition")
	ErrProviderNotLinked = errors.New("provider payment is not linked")
	ErrInvalidMoney      = errors.New("invalid money")
	ErrInvalidNextAction = errors.New("invalid next action")
	ErrPaymentNotFound   = errors.New("payment attempt not found")
	ErrProviderAlreadyLinked  = errors.New("provider payment already linked")
)
