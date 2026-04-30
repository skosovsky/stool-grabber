package telegram

import (
	"context"
	"time"

	"github.com/gotd/td/tgerr"
)

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func transientBackoff(attempt int) time.Duration {
	// простой линейный backoff только для транспортных ошибок
	return time.Duration(200*attempt) * time.Millisecond
}

// invokeRPC выполняет один вызов API: tgerr.FloodWait при необходимости, затем delay_ms после успеха,
// и до maxTransientRPCRetries повторов по ShouldRetryTransientNetwork (без FLOOD_WAIT).
func invokeRPC[T any](ctx context.Context, delayMs int, op func(context.Context) (T, error)) (T, error) {
	var zero T
	transientAttempts := 0
	for {
		res, err := op(ctx)
		if err == nil {
			if throttle := time.Duration(delayMs) * time.Millisecond; throttle > 0 {
				if err := sleepCtx(ctx, throttle); err != nil {
					return zero, err
				}
			}
			return res, nil
		}
		ok, waitErr := tgerr.FloodWait(ctx, err)
		if ok {
			transientAttempts = 0
			continue
		}
		err = waitErr
		if ShouldRetryTransientNetwork(err) && transientAttempts < maxTransientRPCRetries {
			transientAttempts++
			if err := sleepCtx(ctx, transientBackoff(transientAttempts)); err != nil {
				return zero, err
			}
			continue
		}
		return zero, err
	}
}
