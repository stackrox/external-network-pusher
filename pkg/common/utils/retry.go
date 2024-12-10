package utils

import (
	"github.com/cenkalti/backoff/v3"
	"log"
	"time"
)

// WithDefaultRetry retries a given function with a default exponential backoff
func WithDefaultRetry(do func() error) error {
	return WithRetry(do, 2*time.Second, 10*time.Second, 5*time.Minute)
}

// WithRetry retries a given function with an exponential backoff until maxTime is reached
func WithRetry(do func() error, interval, maxInterval, maxTime time.Duration) error {
	exponential := backoff.NewExponentialBackOff()
	exponential.MaxElapsedTime = maxTime
	exponential.InitialInterval = interval
	exponential.MaxInterval = maxInterval

	err := backoff.RetryNotify(do, exponential, func(err error, d time.Duration) {
		log.Printf("call failed, retrying in %s. Error: %v", d.Round(time.Second), err)
	})
	return err
}
