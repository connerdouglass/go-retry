# go-retry

Go library that makes it easy to add automatic retries to your projects, including support for `context.Context`.

## Example with `context.Context`

```go
// Create a context that times out after 10 seconds
ctx := context.Background()
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()

// Perform some action that might fail, with auto retries
err := retry.Run(
    ctx,
    retry.Limit(5), // <-- Limit retries
    retry.Exponential(time.Second), // <-- Exponential backoff
    func(ctx context.Context) error {

        // Do something here:
        err := doSomethingDangerous()

        // If it succeeds:
        return nil

        // If it fails, but is retryable:
        return retry.RetryErr(err)

        // If it fails, and it's not retryable:
        return err

    })
```

## Simpler example

If you don't use `context.Context` or just want something simpler, you can omit the context part entirely:

```go
// Perform some action that might fail, with auto retries
err := retry.Run(
    nil,
    retry.Limit(5), // <-- Limit retries
    retry.Exponential(time.Second), // <-- Exponential backoff
    func(ctx context.Context) error {

        // Do something here:
        err := doSomethingDangerous()

        // If it succeeds:
        return nil

        // If it fails, but is retryable:
        return retry.RetryErr(err)

        // If it fails, and it's not retryable:
        return err

    })
```

## Functional pattern for Delays

In the example above, we use an exponential backoff starting with a 1 second delay. The magnitude of subsequent delays grows exponentially by a factor of two.

However, there are many other choices you could make, depending on your circumstances: `Exponential`, `Fibonacci`, `Linear`, `Constant`. Additionally, this library supports a functional pattern for layering in multiple delay behaviors, such as logging and random offsets.

### Delay with random offset

You can easily add randomized offsets to your Delays with the following functional pattern:

```go
// Just wrap the entire delay in the retry.Rand function
retry.Rand(retry.Exponential(time.Second))
```

This creates a new wrapper Delay that includes the exponential delay, but adds a random offset at the end.

### Adding logging to delays and retries

In much the same way as above, you can add logging so you're notified every time a retry is taking place:

```go
// Logging the plain exponential backoff:
retry.Log(retry.Exponential(time.Second))

// Logging and random offset:
retry.Log(retry.Rand(retry.Exponential(time.Second)))
```

Hopefully you can see that it's very easy to compose clever behaviors with this simple, functional pattern. You can also create your own functions if you choose. Here's the source code for the `Linear` delay function, for an example of how simple it is:

```go
func Linear(base time.Duration) Delay {
	return func(iteration int) time.Duration {
		return base * time.Duration(iteration)
	}
}
```

### Remove all retry limits

If you want no limit on the number of retries, you can use the predefined constant `retry.RetryForever`.

### Running the Example code

There is an example in `example/main.go` that simply fetches the Google homepage using automatic retries. To see the retry behavior in action, turn off your internet connection and run `go run ./example` from the root of this repo.

Try also turning back on your internet access during the middle of the example code's execution. You'll notice it immediately succeeds and ceases retrying.

## Contributions

Contributions are welcome, and encouraged!
