package alarm

import (
	"time"

	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

func (c *Center) RebuildFromSnapshot(rooms map[model.RoomID]differential.PressureRoomState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.notify
	for room, state := range rooms {
		key := "pressure:" + string(room)
		if state.Status == model.PressureAlarm {
			c.records[key] = model.AlarmRecord{
				ID:      key,
				Room:    room,
				Kind:    "pressure",
				Message: "rebuilt pressure alarm from snapshot",
				Level:   severityPressure(state.Status),
				Active:  true,
				At:      time.Now(),
			}
		}
	}
}
