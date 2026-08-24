package differential

import (
	"errors"
	"fmt"

	"cleanroomorcontrol/internal/model"
)

var ErrSnapshotMissing = errors.New("pressure snapshot missing")

type MonitorError struct {
	Room model.RoomID
	Op   string
	Err  error
}

func (e *MonitorError) Error() string {
	return fmt.Sprintf("differential %s room %s: %v", e.Op, e.Room, e.Err)
}

func (e *MonitorError) Unwrap() error {
	return e.Err
}
