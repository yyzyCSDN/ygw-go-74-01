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
	if err := ctx.Err(); err != nil {
		release()
		return &WaitError{Err: ErrInterlockTimeout, Cause: err}
	}
	select {
	case <-settle:
		return nil
	case <-ctx.Done():
		release()
		return &WaitError{Err: ErrInterlockTimeout, Cause: ctx.Err()}
	}
}
