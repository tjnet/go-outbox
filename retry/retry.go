package retry

import (
	"errors"
	"time"
)

type RetryableFunc func() error

type Option func(*config)
type config struct {
	maxAttempts uint
}

type retryableError struct {
	err error
}

// RetryableError marks an error as retryable.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err}
}

// Unwrap implements error wrapping.
func (e *retryableError) Unwrap() error {
	return e.err
}

type unrecoverableError struct {
	error
}

// Unwrap implements error wrapping for unrecoverableError
func (e unrecoverableError) Unwrap() error {
	return e.error
}

// Unrecoverable wraps an error in `unrecoverableError` struct
func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

// Error returns the error string.
func (e *retryableError) Error() string {
	if e.err == nil {
		return "retryable: <nil>"
	}
	return "retryable: " + e.err.Error()
}

// Do -
// this is inspired by
// https://levelup.gitconnected.com/make-your-go-code-more-reliable-with-the-retry-pattern-e9968a2050ba
// https://github.com/avast/retry-go
// https://github.com/thedevsaddam/retry/blob/master/retry.go
func Do(f RetryableFunc, opts ...Option) error {
	config := newDefaultConfig()

	// apply opts
	for _, o := range opts {
		o(config)
	}

	var attempt uint = 0

	for {
		err := f()
		if err == nil {
			return nil
		}

		// Not Retryable
		var rErr *retryableError
		if !errors.As(err, &rErr) {
			return err
		}

		attempt++

		// should stop - exceeded max retry count
		_, shouldStop := next(attempt, config)
		if shouldStop {
			// finds the first error in errors chain that matches target. after that, return unwrapped error
			return rErr.Unwrap()
		}

	}
}

func next(attempt uint, conf *config) (interval time.Duration, shouldStop bool) {
	if attempt >= conf.maxAttempts {
		shouldStop = true
	}
	return 0, shouldStop
}

func WithMaxAttempts(attempts uint) Option {
	// TODO: user builder pattern to meet the requirement
	return func(c *config) {
		c.maxAttempts = attempts
	}
}

func MaxAttempts(attempts uint) Option {
	return func(c *config) {
		c.maxAttempts = attempts
	}
}

func newDefaultConfig() *config {
	return &config{
		maxAttempts: 3,
	}
}
