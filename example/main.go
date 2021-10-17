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

	// Give the operation a few seconds to complete. We also cancel the context is the application
	// receives an interrupt signal (Ctrl+C)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
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

// FetchUrl sends a HTTP GET request to a given URL and returns the response body
func FetchUrl(ctx context.Context, url string) (io.ReadCloser, error) {

	// Send the request to the url
	res, err := http.Get(url)
	if err != nil {
		// Socket / connection errors can always be retried
		return nil, retry.RetryErr(err)
	}

	// 500 errors can always be retried, since they indicate some server-side issue
	if res.StatusCode >= 500 && res.StatusCode < 600 {
		res.Body.Close()
		return nil, retry.RetryErr(fmt.Errorf("http status code: %d", res.StatusCode))
	}

	// If the status code is in the 200 range, the request was successful
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
