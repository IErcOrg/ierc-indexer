package utils

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func WithRetryCount(count int, duration, maxErrDuration time.Duration, fn func() error) func() error {
	return func() error {

		var (
			err       error
			lastErrAt = time.Now()
		)

		for i := 0; i < count; i++ {
			log.Infof("runtime. count: %d", i)
			err = fn()

			if err == nil || errors.Is(err, context.Canceled) {
				return err
			}

			errAt := time.Now()
			errDuration := errAt.Sub(lastErrAt)

			if errDuration > maxErrDuration {
				i = 0
			}

			log.Errorf(
				"error. count: %d. error: %v, lastErrAt: %s, now: %s, errDuration: %s, maxErrDuration: %s",
				i, err, lastErrAt.Format(time.DateTime), errAt.Format(time.DateTime), errDuration, maxErrDuration,
			)

			lastErrAt = errAt
			time.Sleep(duration)
		}

		return err
	}
}
