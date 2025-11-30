package pg

import "errors"

// Sentinel errors returned by this package.
var (
	ErrConnection          = errors.New("connection failed")
	ErrClientClosed        = errors.New("client closed")
	ErrQueryFailed         = errors.New("query execution failed")
	ErrDecodeType          = errors.New("decode row failed")
	ErrEmptyRow            = errors.New("row has 0 columns")
	ErrClientUninitialized = errors.New("client not initialized")
)
