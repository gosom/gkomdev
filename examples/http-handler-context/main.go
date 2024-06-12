package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Register the handler with the middleware, setting the timeout to 5 seconds
	mux.HandleFunc("/", longRunningTaskHandler)

	const timeout = time.Second * 5
	wrappedMux := contextMiddleware(timeout)(mux)

	// Start the server
	http.ListenAndServe(":8080", wrappedMux)
}

// contextMiddleware is a middleware that sets a timeout for the request context.
// If the request takes longer than the timeout, the middleware will cancel the context.
func contextMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set a timeout for the request context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a new request with the updated context
			r = r.WithContext(ctx)

			// Call the next handler with the new context
			next.ServeHTTP(w, r)
		})
	}
}

func longRunningTaskHandler(w http.ResponseWriter, r *http.Request) {
	err := longRunningTask(r.Context())
	if err != nil {
		// if the context expired, return a 504 Gateway Timeout
		if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		// if the task failed for some other reason, return a 500 Internal Server Error
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// if the task completed successfully, return a 200 OK
	w.WriteHeader(http.StatusOK)
}

// longRunningTask is a dummy function that simulates a long running task.
// the task will take 10 seconds to complete.
// If the context is cancelled before the task completes, it will return the error.
func longRunningTask(ctx context.Context) error {
	var dur time.Duration

	// for ~50% of the time, the task will take 10 seconds to complete
	if rand.Float64() > 0.5 {
		dur = 10 * time.Second
	} else {
		dur = 1 * time.Second // and for the other ~50% of the time, the task will take 1 second to complete
	}

	// simulate the task by sleeping for the duration
	select {
	case <-time.After(dur):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
