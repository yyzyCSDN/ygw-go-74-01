package disinfection

import (
	"context"
	"sync"
	"time"

	"cleanroomorcontrol/internal/model"
)

type DoorRelease interface {
	Engage(room model.RoomID) error
	Release(room model.RoomID) error
}

type VentRunner interface {
	Run(ctx context.Context, room model.RoomID) error
}

type Planner struct {
	mu           sync.RWMutex
	phases       map[model.RoomID]model.DisinfectionPhase
	cycles       map[model.RoomID]int
	lastVent     map[model.RoomID]time.Time
	completedAt  map[model.RoomID]time.Time
	ventAttempts map[model.RoomID]int
	schedule     map[model.RoomID][]time.Time
	door         DoorRelease
	vent         VentRunner
}

func NewPlanner(door DoorRelease, vent VentRunner) *Planner {
	return &Planner{
		phases:       make(map[model.RoomID]model.DisinfectionPhase),
		cycles:       make(map[model.RoomID]int),
		lastVent:     make(map[model.RoomID]time.Time),
		completedAt:  make(map[model.RoomID]time.Time),
		ventAttempts: make(map[model.RoomID]int),
		schedule:     make(map[model.RoomID][]time.Time),
		door:         door,
		vent:         vent,
	}
}

func (p *Planner) RegisterRoom(room model.RoomID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phases[room] = model.PhaseIdle
}

func (p *Planner) Start(room model.RoomID) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.phases[room] != model.PhaseIdle {
		return ErrAlreadyDisinfecting
	}
	p.phases[room] = model.PhaseDisinfecting
	if p.door != nil {
		if err := p.door.Engage(room); err != nil {
			return err
		}
	}
	return nil
}

func (p *Planner) CompleteDisinfection(room model.RoomID) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.phases[room] != model.PhaseDisinfecting {
		return ErrNotDisinfecting
	}
	p.phases[room] = model.PhaseVentilating
	return nil
}

func (p *Planner) CompleteVentilation(room model.RoomID) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.phases[room] != model.PhaseVentilating {
		return ErrNotVentilating
	}
	p.phases[room] = model.PhaseIdle
	p.cycles[room]++
	if p.lastVent[room].IsZero() {
		p.lastVent[room] = time.Now()
	}
	p.completedAt[room] = time.Now()
	p.ventAttempts[room] = 0
	p.schedule[room] = nil
	if p.door != nil {
		if err := p.door.Release(room); err != nil {
			return err
		}
	}
	return nil
}

func (p *Planner) Phase(room model.RoomID) model.DisinfectionPhase {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.phases[room]
}

func (p *Planner) Cycles(room model.RoomID) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cycles[room]
}

func (p *Planner) LastVent(room model.RoomID) (time.Time, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	at, ok := p.lastVent[room]
	return at, ok
}

func (p *Planner) CompletedAt(room model.RoomID) (time.Time, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	at, ok := p.completedAt[room]
	return at, ok
}

func (p *Planner) Phases() map[model.RoomID]model.DisinfectionPhase {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[model.RoomID]model.DisinfectionPhase, len(p.phases))
	for room, phase := range p.phases {
		out[room] = phase
	}
	return out
}
