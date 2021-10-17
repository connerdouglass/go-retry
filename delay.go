package retry

import (
	cryptoRand "crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	mathRand "math/rand"
	"os"
	"time"
)

// Delay defines a backoff / delay strategy for retrying
type Delay func(iteration int) time.Duration

// Exponential creates an exponential backoff delay, which uses the following sequence:
// [ 1, 2, 4, 8, 16, ... ] times the `base` duration
func Exponential(base time.Duration) Delay {
	return func(iteration int) time.Duration {
		sleepFactor := 1 << (iteration - 1)
		return time.Second * time.Duration(sleepFactor)
	}
}

// Linear creates a linear backoff delay, which uses the following sequence:
// [ 1, 2, 3, 4, 5, ... ] times the `base` duration
func Linear(base time.Duration) Delay {
	return func(iteration int) time.Duration {
		return base * time.Duration(iteration)
	}
}

// Constant creates a constant backoff delay, which always delays the exact same duration:
// [ 1, 1, 1, 1, 1, ... ] times the `base` duration
func Constant(delay time.Duration) Delay {
	return func(iteration int) time.Duration {
		return delay
	}
}

// NoDelay creates a delay strategy that doesn't delay at all
func NoDelay() Delay {
	return Constant(0)
}

// Fibonacci creates a fibonacci backoff delay, which uses the Fibonacci sequence:
// [ 1, 1, 2, 3, 5, 8 ... ] times the `base` duration.
// This function is stateful, so should not be used multiple times. Instead a new instance
// of Fibonacci should be called for each use of Retry
func Fibonacci(base time.Duration) Delay {
	beforePrevious, previous := 1, 1
	fib := func(iteration int) int {
		if iteration <= 2 {
			return 1
		}
		next := beforePrevious + previous
		beforePrevious = previous
		previous = next
		return next
	}
	return func(iteration int) time.Duration {
		return base * time.Duration(fib(iteration))
	}
}

// Rand takes another delay and applies a randomized offset to each value with a magnitude between 0 and 10% of the original
// and the offset is random for each iteration of the delay function
func Rand(delay Delay) Delay {
	return func(iteration int) time.Duration {
		originalSleepDuration := delay(iteration)
		maxOffset := originalSleepDuration / 10
		return originalSleepDuration + randDuration(maxOffset)
	}
}

// Log adds a logging layer to a delay, which logs a message whenever a delay is taking place
func Log(delay Delay) Delay {
	return LogWithOptions(
		delay,
		os.Stdout,
		func(sleepDuration time.Duration) string {
			return fmt.Sprintf("Sleeping %s then retrying", sleepDuration)
		},
	)
}

func LogWithOptions(delay Delay, out io.Writer, formatter func(time.Duration) string) Delay {
	return func(iteration int) time.Duration {
		sleepDuration := delay(iteration)
		message := formatter(sleepDuration)
		out.Write([]byte(message + "\n"))
		return sleepDuration
	}
}

func randPositiveInt64() int64 {
	var randInt int64

	// Attempt to get the random value using crypto/rand. If there's an error, use math/rand
	randBigInt, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		randInt = mathRand.Int63()
	} else {
		randInt = randBigInt.Int64()
	}

	// Ensure the result is always positive
	if randInt < 0 {
		return -randInt
	}

	return randInt
}

func randDuration(max time.Duration) time.Duration {
	randFloatZeroToOne := float64(randPositiveInt64()) / float64(math.MaxInt64)
	return time.Duration(float64(max) * (randFloatZeroToOne*2 - 1))
}
