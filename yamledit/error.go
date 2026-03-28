package yamledit

import "errors"

// ErrNoop is a sentinel error that indicates no operation should be performed.
var ErrNoop = errors.New("noop")
