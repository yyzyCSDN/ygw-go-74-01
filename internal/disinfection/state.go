package disinfection

import "cleanroomorcontrol/internal/model"

func PhaseMap(phases map[model.RoomID]model.DisinfectionPhase) map[model.RoomID]string {
	out := make(map[model.RoomID]string, len(phases))
	for room, phase := range phases {
		out[room] = phase.String()
	}
	return out
}
