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

// waitForSettle blocks until airflow for the room reports a settle signal on
// settle, or ctx is cancelled (for example by the 30-second interlock timeout).
// On cancellation it runs release to undo the acquisition, so the interlock is
// restored to available instead of staying occupied after a timeout.
func waitForSettle(ctx context.Context, settle <-chan struct{}, release func()) error {
	if release == nil {
		release = func() {}
	}
	select {
	case <-settle:
		return nil
	case <-ctx.Done():
		release()
		cause := ctx.Err()
		if cause == nil {
			cause = context.DeadlineExceeded
		}
		return &WaitError{Err: ErrInterlockTimeout, Cause: cause}
	}
}
