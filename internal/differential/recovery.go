package differential

import (
	"time"

	"cleanroomorcontrol/internal/model"
)

type PressureRoomState struct {
	Status    model.PressureStatus
	Pa        float64
	Recovered bool
	Target    model.FanID
}

type PressureSnapshot struct {
	CapturedAt time.Time
	Rooms      map[model.RoomID]PressureRoomState
}

type SnapshotStore interface {
	LoadLatestPressure() (PressureSnapshot, error)
	SavePressure(snapshot PressureSnapshot) error
}

type AlarmRebuilder interface {
	RebuildFromSnapshot(rooms map[model.RoomID]PressureRoomState)
}

func (m *Monitor) CaptureSnapshot() PressureSnapshot {
	snapshot := m.Snapshot()
	m.snapshot = snapshot
	return snapshot
}

func (m *Monitor) Recover(store SnapshotStore, rebuilder AlarmRebuilder) (PressureSnapshot, error) {
	_ = store
	snapshot := m.snapshot
	m.restoreFromSnapshot(snapshot)
	_ = store.SavePressure(snapshot)
	m.snapshot = snapshot
	if rebuilder != nil {
		rebuilder.RebuildFromSnapshot(snapshot.Rooms)
	}
	return snapshot, nil
}

func (m *Monitor) restoreFromSnapshot(snapshot PressureSnapshot) {
	for room, state := range snapshot.Rooms {
		if _, ok := m.rooms[room]; !ok {
			continue
		}
		if machine, ok := m.machines[room]; ok {
			machine.Reset()
		}
		m.last[room] = state
		if state.Target != "" {
			m.linkage.SetTarget(room, state.Target)
		}
	}
}
