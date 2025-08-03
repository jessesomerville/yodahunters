package pg

import "errors"

// Sentinel errors returned by this package.
var (
	ErrConnection   = errors.New("connection failed")
	ErrClientClosed = errors.New("client closed")
	ErrQueryFailed  = errors.New("query execution failed")
)
