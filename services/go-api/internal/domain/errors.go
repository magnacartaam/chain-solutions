package domain

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrSessionInactive   = errors.New("session is inactive")
	ErrInvalidSeed       = errors.New("invalid client seed")
	ErrBatchClosed       = errors.New("batch is already closed")
)
