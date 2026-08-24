package differential

import (
	"sync"

	"cleanroomorcontrol/internal/model"
)

type Linkage struct {
	mu       sync.RWMutex
	targets  map[model.RoomID]model.FanID
	pending  map[model.RoomID]model.FanID
	revision map[model.RoomID]int
	history  map[model.RoomID][]model.FanID
}

func NewLinkage() *Linkage {
	return &Linkage{
		targets:  make(map[model.RoomID]model.FanID),
		pending:  make(map[model.RoomID]model.FanID),
		revision: make(map[model.RoomID]int),
		history:  make(map[model.RoomID][]model.FanID),
	}
}

func (l *Linkage) SetTarget(room model.RoomID, fan model.FanID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.targets[room] = fan
	l.pending[room] = fan
	l.revision[room]++
}

func (l *Linkage) ApplyFanSwitch(room model.RoomID, fan model.FanID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if fan == "" {
		return
	}
	l.pending[room] = fan
	l.revision[room]++
}

func (l *Linkage) Target(room model.RoomID) (model.FanID, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	fan, ok := l.targets[room]
	return fan, ok
}

func (l *Linkage) Revision(room model.RoomID) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.revision[room]
}

func (l *Linkage) History(room model.RoomID) []model.FanID {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]model.FanID, len(l.history[room]))
	copy(out, l.history[room])
	return out
}

func (l *Linkage) Snapshot() map[model.RoomID]model.FanID {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[model.RoomID]model.FanID, len(l.targets))
	for room, fan := range l.targets {
		out[room] = fan
	}
	return out
}
