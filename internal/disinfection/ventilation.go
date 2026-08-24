package disinfection

import (
	"context"

	"cleanroomorcontrol/internal/model"
)

func (p *Planner) Ventilate(ctx context.Context, room model.RoomID) error {
	p.mu.Lock()
	phase, registered := p.phases[room]
	p.ventAttempts[room]++
	p.mu.Unlock()
	if !registered {
		return ErrNotVentilating
	}
	if phase != model.PhaseVentilating {
		return ErrNotVentilating
	}
	if err := ctx.Err(); err != nil {
		return ErrVentilationTimeout
	}
	if err := p.vent.Run(ctx, room); err != nil {
		return err
	}
	return p.CompleteVentilation(room)
}

func (p *Planner) VentAttempts(room model.RoomID) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ventAttempts[room]
}
