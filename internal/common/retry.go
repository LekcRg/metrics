package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func Retry(ctx context.Context, rfunc func() error) error {
	timeouts := []int{1, 3, 5}
	var err error
	for i := range 4 {
		err = rfunc()

		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		var netErr net.Error

		isRetryable := (errors.Is(err, context.DeadlineExceeded) ||
			errors.As(err, &netErr) ||
			(errors.As(err, &pgErr) &&
				(pgErr.Code == pgerrcode.ConnectionException ||
					pgErr.Code == pgerrcode.DeadlockDetected ||
					pgerrcode.IsConnectionException(pgErr.Code))))

		if !isRetryable || i == 3 {
			return err
		}

		logText := fmt.Sprintf("repeat attempt %d after %d seconds", i+1, timeouts[i])
		logger.Log.Warn(logText)

		timer := time.NewTimer(time.Duration(timeouts[i]) * time.Second)

		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("stopped")
		case <-timer.C:
		}
	}

	return err
}
