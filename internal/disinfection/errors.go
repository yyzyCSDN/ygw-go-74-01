package disinfection

import "errors"

var ErrAlreadyDisinfecting = errors.New("room is already disinfecting")

var ErrNotDisinfecting = errors.New("room is not in disinfecting phase")

var ErrNotVentilating = errors.New("room is not in ventilating phase")

var ErrVentilationTimeout = errors.New("ventilation timed out")
