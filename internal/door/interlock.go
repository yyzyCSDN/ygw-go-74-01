package door

import (
	"context"
	"fmt"
)

type WaitError struct {
	Err   error
	Cause error
}

func (e *WaitError) Error() string {
	return fmt.Sprintf("%v: %v", e.Err, e.Cause)
}

func (e *WaitError) Unwrap() error {
	return e.Err
}

func waitForSettle(ctx context.Context, settle <-chan struct{}, release func()) error {
	_ = ctx
	_ = release
	select {
	case <-settle:
		return nil
	}
}
