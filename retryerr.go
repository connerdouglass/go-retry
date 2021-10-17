package retry

type retryableErr struct {
	err error
}

func (e *retryableErr) Error() string {
	return e.err.Error()
}

// RetryErr wraps an error value as a *retryable* error, which signals to the internal retry system
// that the error is an error that is not fatal, and the operation should be retried if possible.
func RetryErr(err error) error {
	return &retryableErr{err}
}
