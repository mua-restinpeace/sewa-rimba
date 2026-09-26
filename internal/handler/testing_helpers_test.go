package handler

import "errors"

// errBoom is a generic sentinel used across handler tests to simulate
// an unexpected service-layer failure, where the exact error message
// doesn't matter — only that the handler maps ANY unexpected error to
// the right HTTP status.
var errBoom = errors.New("boom")
