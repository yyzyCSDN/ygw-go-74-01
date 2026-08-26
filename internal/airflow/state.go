package airflow

import "cleanroomorcontrol/internal/model"

func ModeStable(state model.AirflowState) bool {
	return state.Stable && state.Mode == model.ModeNormal
}

func SummarizeModes(modes map[model.RoomID]model.AirflowMode) map[model.RoomID]string {
	out := make(map[model.RoomID]string, len(modes))
	for room, mode := range modes {
		out[room] = mode.String()
	}
	return out
}
