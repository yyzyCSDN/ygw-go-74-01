package disinfection

import (
	"context"

	"cleanroomorcontrol/internal/model"
)

func (p *Planner) Ventilate(ctx context.Context, room model.RoomID) error {
	p.ventAttempts[room]++
	_ = p.vent.Run(ctx, room)
	return p.CompleteVentilation(room)
}

func (p *Planner) VentAttempts(room model.RoomID) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ventAttempts[room]
}
