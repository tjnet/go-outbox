package retry

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestDoSuccessFully(t *testing.T) {
	var expectedErr error = nil

	err := Do(func() error {
		return nil
	})
	assert.Equal(t, expectedErr, err)
}

func TestExitOnMaxAttempt(t *testing.T) {
	attempts := 0
	err := Do(func() error {
		attempts++
		return RetryableError(fmt.Errorf("this is retryable error"))
	}, WithMaxAttempts(10))

	assert.NotNil(t, err, "expected error")
	assert.Equal(t, attempts, 10)
}

func TestUnWraps(t *testing.T) {
	err := Do(func() error {
		return RetryableError(io.EOF)
	})

	assert.NotNil(t, err, "expected error")
	assert.Equal(t, err, io.EOF)
}

func TestUnRecoverableError(t *testing.T) {
	attempts := 0
	testErr := errors.New("error")

	err := Do(func() error {
		attempts++
		return Unrecoverable(testErr)
	})
	assert.Equal(t, testErr, errors.Unwrap(err))
	assert.Equal(t, 1, attempts, "unrecoverable error broke the loop")
}

func TestExampleDo(t *testing.T) {
	i := 0
	if err := Do(func() error {
		fmt.Printf("%d\n", i)
		i++
		return RetryableError(fmt.Errorf("oops"))
	}, WithMaxAttempts(3)); err != nil {
		// handle error
		fmt.Printf("error occured: %v\n", err)
	}
}
