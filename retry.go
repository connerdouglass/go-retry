package retry

import (
	"context"
	"time"
)

// RetryFunc defines the function signature for functions that can be retried
type RetryFunc func(ctx context.Context) error

// // Run uses the default options to call and then retry a function, using an exponential backoff with the
// // given base delay, and limited to a specific number of retries
// func Run(
// 	ctx context.Context,
// 	retryLimit Limit,
// 	baseDelay time.Duration,
// 	fn RetryFunc,
// ) error {
// 	return RunWithOptions(
// 		ctx,
// 		retryLimit,
// 		ExponentialBackoff(baseDelay),
// 		fn,
// 	)
// }

// RunWithContext runs a retryable function, retrying it whenever an eligible error is encountered,
// until the maximum number of retries has been met, at which point the latest non-nil error
// is returned.
func Run(
	ctx context.Context,
	retryLimit Limit,
	delay Delay,
	fn RetryFunc,
) error {

	// Ensure there is some context, if nil is passed.
	if ctx == nil {
		ctx = context.Background()
	}

	// Track the latest error
	var lastError error

	// Loop until our limiter runs out. Note that if the retryLimit is negative, it runs forever
	for iteration := 0; retryLimit < 0 || Limit(iteration) <= retryLimit; iteration++ {

		// If this is not the first iteration, we need to sleep before retrying
		if iteration > 0 {
			sleepDuration := delay(iteration)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
				break
			}
		}

		// Run the function and store the error
		err := fn(ctx)
		shouldRetry := isErrRetryable(err)
		if err != nil {
			lastError = err
		}

		// If we should retry, retry
		if shouldRetry {
			continue
		}

		// If we should not retry, return the error (it might be nil in the case of success)
		return err

	}

	// If we get here, we exceeded the maximum number of retries
	if lastError == nil {
		lastError = ErrTooManyRetries
	} else if err, ok := lastError.(*retryableErr); ok {
		// Pull out the wrapped raw error instance
		lastError = err.err
	}
	return lastError

}

func isErrRetryable(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*retryableErr)
	return ok
}
