package disinfection

import (
	"time"

	"cleanroomorcontrol/internal/model"
)

func (p *Planner) Schedule(room model.RoomID, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.schedule[room] = append(p.schedule[room], at)
}

func (p *Planner) Due(room model.RoomID, now time.Time) []time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var due []time.Time
	for _, at := range p.schedule[room] {
		if !at.After(now) {
			due = append(due, at)
		}
	}
	return due
}

func (p *Planner) NextDue(room model.RoomID, now time.Time) (time.Time, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var next time.Time
	found := false
	for _, at := range p.schedule[room] {
		if at.After(now) && (!found || at.Before(next)) {
			next = at
			found = true
		}
	}
	return next, found
}
