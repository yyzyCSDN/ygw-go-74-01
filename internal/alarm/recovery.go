package alarm

import (
	"time"

	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

func (c *Center) RebuildFromSnapshot(rooms map[model.RoomID]differential.PressureRoomState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for room, state := range rooms {
		key := "pressure:" + string(room)
		if state.Status == model.PressureAlarm && !state.Recovered {
			c.records[key] = model.AlarmRecord{
				ID:      key,
				Room:    room,
				Kind:    "pressure",
				Message: "rebuilt pressure alarm from snapshot",
				Level:   severityPressure(state.Status),
				Active:  true,
				At:      time.Now(),
			}
			continue
		}
		if record, ok := c.records[key]; ok && record.Active {
			record.Active = false
			c.records[key] = record
		}
	}
}
