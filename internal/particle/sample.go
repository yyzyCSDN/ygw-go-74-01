package particle

import (
	"time"

	"cleanroomorcontrol/internal/model"
)

type CycleResult struct {
	Room    model.RoomID
	Samples int
	Errors  int
	At      time.Time
}

func (m *Monitor) RunCycle(samples []model.ParticleSample) CycleResult {
	result := CycleResult{At: time.Now()}
	for _, sample := range samples {
		result.Samples++
		if sample.Room != "" {
			result.Room = sample.Room
		}
		if err := m.WriteSample(sample); err != nil {
			result.Errors++
		}
	}
	return result
}

func (m *Monitor) RotateAll(rooms []model.RoomID) []error {
	var errs []error
	for _, room := range rooms {
		if err := m.RotateRecord(room); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
