package differential

import (
	"time"

	"cleanroomorcontrol/internal/model"
)

type AlarmSink interface {
	ReportPressure(room model.RoomID, status model.PressureStatus, pa float64)
}

type FanStateProvider interface {
	FanState(fan model.FanID) model.FanState
}

type Monitor struct {
	rooms    map[model.RoomID]model.Room
	machines map[model.RoomID]*PressureStateMachine
	linkage  *Linkage
	sink     AlarmSink
	fans     FanStateProvider
	last     map[model.RoomID]PressureRoomState
	snapshot PressureSnapshot
}

func NewMonitor(rooms []model.Room, linkage *Linkage, sink AlarmSink, fans FanStateProvider) *Monitor {
	monitor := &Monitor{
		rooms:    make(map[model.RoomID]model.Room),
		machines: make(map[model.RoomID]*PressureStateMachine),
		linkage:  linkage,
		sink:     sink,
		fans:     fans,
		last:     make(map[model.RoomID]PressureRoomState),
	}
	for _, room := range rooms {
		monitor.RegisterRoom(room)
	}
	return monitor
}

func (m *Monitor) RegisterRoom(room model.Room) {
	m.rooms[room.ID] = room
	m.machines[room.ID] = NewPressureStateMachine(room.PressureLow, room.PressureCritical, 3, 2)
	m.last[room.ID] = PressureRoomState{Status: model.PressureStable}
}

func (m *Monitor) Sample(reading model.PressureReading) model.PressureStatus {
	if _, ok := m.rooms[reading.Room]; !ok {
		return model.PressureStable
	}
	machine := m.machines[reading.Room]
	status := machine.Feed(reading.Pa)
	state := m.last[reading.Room]
	state.Pa = reading.Pa
	state.Status = status
	m.last[reading.Room] = state
	if status == model.PressureAlarm && m.sink != nil {
		m.sink.ReportPressure(reading.Room, status, reading.Pa)
	}
	return status
}

func (m *Monitor) Status(room model.RoomID) model.PressureStatus {
	machine, ok := m.machines[room]
	if !ok {
		return model.PressureStable
	}
	return machine.Status()
}

func (m *Monitor) LastPa(room model.RoomID) float64 {
	state, ok := m.last[room]
	if !ok {
		return 0
	}
	return state.Pa
}

func (m *Monitor) LinkageTarget(room model.RoomID) (model.FanID, bool) {
	return m.linkage.Target(room)
}

func (m *Monitor) MaintenanceTarget(room model.RoomID) model.FanID {
	target, ok := m.linkage.Target(room)
	if !ok {
		return ""
	}
	if m.linkage.Revision(room) < 1 {
		return ""
	}
	if m.fans != nil {
		state := m.fans.FanState(target)
		if state == model.FanFailed {
			return ""
		}
		if state == model.FanStopped {
			return ""
		}
		if state != model.FanRunning {
			return ""
		}
	}
	return target
}

func (m *Monitor) LinkageHistory(room model.RoomID) []model.FanID {
	return m.linkage.History(room)
}

func (m *Monitor) Snapshot() PressureSnapshot {
	rooms := make(map[model.RoomID]PressureRoomState, len(m.last))
	for room, state := range m.last {
		target, _ := m.linkage.Target(room)
		state.Target = target
		rooms[room] = state
	}
	return PressureSnapshot{CapturedAt: time.Now(), Rooms: rooms}
}
