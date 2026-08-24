package env

import (
	"sync"

	"cleanroomorcontrol/internal/model"
)

type EnvAlarmSink interface {
	ReportEnv(room model.RoomID, reading model.EnvReading, violations []string)
}

type Monitor struct {
	mu     sync.RWMutex
	limits *LimitsStore
	last   map[model.RoomID]model.EnvReading
	sink   EnvAlarmSink
}

func NewMonitor(limits *LimitsStore, sink EnvAlarmSink) *Monitor {
	return &Monitor{
		limits: limits,
		last:   make(map[model.RoomID]model.EnvReading),
		sink:   sink,
	}
}

func (m *Monitor) Sample(reading model.EnvReading) []string {
	normalized := reading.Normalized()
	limits, ok := m.limits.Get(normalized.Room)
	m.mu.Lock()
	m.last[normalized.Room] = normalized
	m.mu.Unlock()
	if !ok {
		return nil
	}
	violations := limits.Violations(normalized)
	if len(violations) > 0 && m.sink != nil {
		m.sink.ReportEnv(normalized.Room, normalized, violations)
	}
	return violations
}

func (m *Monitor) Latest(room model.RoomID) (model.EnvReading, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	reading, ok := m.last[room]
	return reading, ok
}

func (m *Monitor) All() map[model.RoomID]model.EnvReading {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[model.RoomID]model.EnvReading, len(m.last))
	for room, reading := range m.last {
		out[room] = reading
	}
	return out
}
