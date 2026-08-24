package airflow

import (
	"errors"
	"fmt"

	"cleanroomorcontrol/internal/model"
)

var ErrUnknownFan = errors.New("unknown fan unit")

var ErrStabilityTimeout = errors.New("airflow stability wait timed out")

var ErrNoSignalSource = errors.New("airflow signal source is missing")

var ErrInvalidModeTransition = errors.New("invalid airflow mode transition")

var ErrFanStartFailed = errors.New("fan start failed")

type SwitchError struct {
	Room model.RoomID
	Fan  model.FanID
	Op   string
	Err  error
}

func (e *SwitchError) Error() string {
	return fmt.Sprintf("airflow %s room %s fan %s: %v", e.Op, e.Room, e.Fan, e.Err)
}

func (e *SwitchError) Unwrap() error {
	return e.Err
}
