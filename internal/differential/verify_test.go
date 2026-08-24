package differential_test

import (
	"context"
	"testing"

	"cleanroomorcontrol/internal/airflow"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

type probeDriver struct{}

func (probeDriver) Start(fan model.FanID) error { return nil }

func (probeDriver) Stop(fan model.FanID) error { return nil }

type probeSignals struct{}

func (probeSignals) SettleChannel(room model.RoomID) <-chan struct{} {
	return make(chan struct{})
}

func TestDifferentialRefreshOnFanSwitch(t *testing.T) {
	room := model.Room{ID: "OR-1", PressureTarget: 25, PressureLow: 15, PressureCritical: 5}
	linkage := differential.NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	monitor := differential.NewMonitor([]model.Room{room}, linkage, nil, nil)
	units := []model.FanUnit{
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning},
		{ID: "EX-2", Role: model.FanExhaust, State: model.FanStandby},
	}
	controller := airflow.NewController(units, probeDriver{}, linkage, probeSignals{})
	controller.RegisterRoom("OR-1")
	if err := controller.SwitchExhaustFan(context.Background(), "OR-1", "EX-2"); err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	target, ok := monitor.LinkageTarget("OR-1")
	if !ok || target != "EX-2" {
		t.Fatalf("linkage target = %q ok=%v, want EX-2", target, ok)
	}
	if revision := linkage.Revision("OR-1"); revision < 2 {
		t.Fatalf("revision = %d, want >= 2", revision)
	}
	status := monitor.Sample(model.PressureReading{Room: "OR-1", Pa: 8})
	if status == model.PressureAlarm {
		t.Fatal("linkage target is unusable, pressure maintenance broken")
	}
}
