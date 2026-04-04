package pg

import "errors"

// Sentinel errors returned by this package.
var (
	ErrConnection          = errors.New("connection failed")
	ErrClientUninitialized = errors.New("client not initialized")
)
