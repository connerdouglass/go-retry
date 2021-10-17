package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/connerdouglass/go-retry"
)

func main() {

	// Give the operation 10 seconds to complete. We also cancel the context is the application is killed
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	// Fetch the url, with automatic retry
	var response io.ReadCloser
	err := retry.Run(
		ctx,
		retry.Limit(5),
		retry.Log(retry.Rand(retry.Exponential(time.Second))),
		func(ctx context.Context) error {
			res, err := FetchUrl(ctx, "https://google.com")
			if err != nil {
				return err
			}
			response = res
			return nil
		})

	// Log the results
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		io.Copy(os.Stdout, response)
		response.Close()
	}

}

func FetchUrl(ctx context.Context, url string) (io.ReadCloser, error) {

	// Send the request to the url
	res, err := http.Get(url)
	if err != nil {
		return nil, retry.RetryErr(err) // socket errors can always be retried
	}

	// 500 errors can always be retried
	if res.StatusCode >= 500 && res.StatusCode < 600 {
		res.Body.Close()
		return nil, retry.RetryErr(fmt.Errorf("http status code: %d", res.StatusCode))
	}

	// If the error is in the 200 range
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if err != nil {
			return nil, err
		}
		return res.Body, nil
	}

	// Other errors cannot be retried
	res.Body.Close()
	return nil, fmt.Errorf("http status code: %d", res.StatusCode)

}
