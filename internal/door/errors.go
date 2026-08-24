package door

import "errors"

var ErrDoorMissing = errors.New("door not registered")

var ErrDoorBusy = errors.New("door is busy")

var ErrInterlockTimeout = errors.New("interlock wait timed out")
