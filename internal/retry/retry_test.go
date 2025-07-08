package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// Mock для net.Error
type mockNetError struct {
	msg       string
	timeout   bool
	temporary bool
}

func (e *mockNetError) Error() string   { return e.msg }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "network timeout error",
			err:      &mockNetError{timeout: true, msg: "timeout"},
			expected: true,
		},
		{
			name:     "network non-timeout error",
			err:      &mockNetError{timeout: false, msg: "non-timeout"},
			expected: false,
		},
		{
			name:     "postgres connection exception",
			err:      &pgconn.PgError{Code: pgerrcode.ConnectionException},
			expected: true,
		},
		{
			name:     "postgres deadlock detected",
			err:      &pgconn.PgError{Code: pgerrcode.DeadlockDetected},
			expected: true,
		},
		{
			name:     "postgres connection failure",
			err:      &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			expected: true,
		},
		{
			name:     "postgres syntax error",
			err:      &pgconn.PgError{Code: pgerrcode.SyntaxError},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	rfunc := func() error {
		callCount++
		return nil
	}

	err := Retry(ctx, rfunc)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	rfunc := func() error {
		callCount++
		if callCount < 3 {
			return context.DeadlineExceeded // retryable error
		}
		return nil
	}

	err := Retry(ctx, rfunc)

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	expectedErr := errors.New("non-retryable error")

	rfunc := func() error {
		callCount++
		return expectedErr
	}

	err := Retry(ctx, rfunc)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_MaxAttemptsReached(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	expectedErr := context.DeadlineExceeded // retryable error

	rfunc := func() error {
		callCount++
		return expectedErr
	}

	err := Retry(ctx, rfunc)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 4, callCount) // 1 initial + 3 retries
}
