package differential_test

import (
	"testing"
	"time"

	"cleanroomorcontrol/internal/alarm"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type probeSnapshotStore struct {
	snapshot differential.PressureSnapshot
}

func (s *probeSnapshotStore) LoadLatestPressure() (differential.PressureSnapshot, error) {
	return s.snapshot, nil
}

func (s *probeSnapshotStore) SavePressure(snapshot differential.PressureSnapshot) error {
	s.snapshot = snapshot
	return nil
}

func TestDifferentialRecoveryUsesLatestSnapshot(t *testing.T) {
	room := model.Room{ID: "OR-1", PressureTarget: 25, PressureLow: 15, PressureCritical: 5}
	linkage := differential.NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	store := &probeSnapshotStore{}
	monitor := differential.NewMonitor([]model.Room{room}, linkage, nil, nil)
	monitor.Sample(model.PressureReading{Room: "OR-1", Pa: 10})
	monitor.Sample(model.PressureReading{Room: "OR-1", Pa: 10})
	monitor.Sample(model.PressureReading{Room: "OR-1", Pa: 10})
	monitor.Sample(model.PressureReading{Room: "OR-1", Pa: 3})
	monitor.CaptureSnapshot()
	store.SavePressure(differential.PressureSnapshot{
		CapturedAt: time.Now(),
		Rooms: map[model.RoomID]differential.PressureRoomState{
			"OR-1": {Status: model.PressureStable, Recovered: true, Target: "EX-1"},
		},
	})
	center := alarm.NewCenter(alarm.NewCache(particle.NewTable()), nil)
	if _, err := monitor.Recover(store, center); err != nil {
		t.Fatalf("recover failed: %v", err)
	}
	if center.HasActive("pressure:OR-1") {
		t.Fatal("recovered pressure alarm must not be re-raised after restart")
	}
}
