package retry

import "errors"

// ErrTooManyRetries is the standard error that is returned when there have been too many retries
// within a run. There is one global instance of this error, so it can be easily checked for
var ErrTooManyRetries = errors.New("exceeded retry limit")
