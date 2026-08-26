package door

import "cleanroomorcontrol/internal/model"

func SummarizeDoors(states map[model.RoomID]model.DoorState) map[model.RoomID]string {
	out := make(map[model.RoomID]string, len(states))
	for room, state := range states {
		out[room] = state.String()
	}
	return out
}

func AllowedEntry(states map[model.RoomID]model.DoorState) []model.RoomID {
	var rooms []model.RoomID
	for room, state := range states {
		if model.CanEnterDoor(state) {
			rooms = append(rooms, room)
		}
	}
	return rooms
}
