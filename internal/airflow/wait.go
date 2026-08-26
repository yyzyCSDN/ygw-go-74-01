package airflow

import (
	"context"

	"cleanroomorcontrol/internal/model"
)

type SignalSource interface {
	SettleChannel(room model.RoomID) <-chan struct{}
}

type WaitError struct {
	Room  model.RoomID
	Err   error
	Cause error
}

func (e *WaitError) Error() string {
	if e.Cause != nil {
		return "airflow wait room " + string(e.Room) + ": " + e.Err.Error() + ": " + e.Cause.Error()
	}
	return "airflow wait room " + string(e.Room) + ": " + e.Err.Error()
}

func (e *WaitError) Unwrap() error {
	return e.Err
}

func WaitStable(ctx context.Context, room model.RoomID, source SignalSource) error {
	if source == nil {
		return &WaitError{Room: room, Err: ErrNoSignalSource}
	}
	select {
	case <-source.SettleChannel(room):
		if ctx.Err() != nil {
			return &WaitError{Room: room, Err: ErrStabilityTimeout, Cause: ctx.Err()}
		}
		return nil
	case <-ctx.Done():
		cause := ctx.Err()
		if cause == nil {
			cause = context.DeadlineExceeded
		}
		return &WaitError{Room: room, Err: ErrStabilityTimeout, Cause: cause}
	}
}
