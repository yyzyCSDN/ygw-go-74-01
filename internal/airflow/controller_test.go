package airflow

import (
	"testing"

	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

// stubDriver is a minimal FanDriver for tests; it never fails and records calls.
type stubDriver struct {
	started map[model.FanID]bool
	stopped map[model.FanID]bool
}

func newStubDriver() *stubDriver {
	return &stubDriver{started: make(map[model.FanID]bool), stopped: make(map[model.FanID]bool)}
}

func (d *stubDriver) Start(fan model.FanID) error { d.started[fan] = true; return nil }
func (d *stubDriver) Stop(fan model.FanID) error  { d.stopped[fan] = true; return nil }

func newSwitchController() (*Controller, *differential.Linkage, *stubDriver) {
	units := []model.FanUnit{
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning, Airflow: 3600},
		{ID: "EX-2", Role: model.FanExhaust, State: model.FanStandby, Airflow: 3550},
	}
	linkage := differential.NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	driver := newStubDriver()
	controller := NewController(units, driver, linkage, nil)
	controller.RegisterRoom("OR-1")
	return controller, linkage, driver
}

// TestSwitchExhaustFanRefreshesLinkageTarget reproduces the reported defect:
// after switching to the standby exhaust unit the linkage target must follow
// the newly running fan, otherwise pressure-differential linkage stays pinned
// to the (now stopped) previous unit.
func TestSwitchExhaustFanRefreshesLinkageTarget(t *testing.T) {
	controller, linkage, driver := newSwitchController()

	if err := controller.SwitchExhaustFan(nil, "OR-1", "EX-2"); err != nil {
		t.Fatalf("switch exhaust fan: %v", err)
	}

	target, ok := linkage.Target("OR-1")
	if !ok || target != "EX-2" {
		t.Fatalf("linkage target after switch = %q, want EX-2", target)
	}

	// The new unit must be running and the maintenance target must resolve to it
	// rather than the previous unit.
	if got := controller.FanState("EX-2"); got != model.FanRunning {
		t.Fatalf("EX-2 state = %v, want running", got)
	}
	if got := controller.FanState("EX-1"); got != model.FanStopped {
		t.Fatalf("EX-1 state = %v, want stopped", got)
	}

	// The driver must have been asked to stop the old unit and start the new one.
	if !driver.stopped["EX-1"] {
		t.Fatal("old exhaust unit was not stopped")
	}
	if !driver.started["EX-2"] {
		t.Fatal("new exhaust unit was not started")
	}

	// currentExhaust must resolve to the new unit via the refreshed linkage.
	if got := controller.currentExhaust("OR-1"); got != "EX-2" {
		t.Fatalf("currentExhaust = %q, want EX-2", got)
	}

	// Switch events must record the real from->to transition.
	events := controller.Events()
	if len(events) == 0 || events[len(events)-1].From != "EX-1" || events[len(events)-1].To != "EX-2" {
		t.Fatalf("switch events = %+v", events)
	}
}

// TestSwitchExhaustFanToSameUnitKeepsTarget guards against marking the only
// running unit as stopped when "switching" to itself.
func TestSwitchExhaustFanToSameUnitKeepsTarget(t *testing.T) {
	controller, linkage, _ := newSwitchController()

	if err := controller.SwitchExhaustFan(nil, "OR-1", "EX-1"); err != nil {
		t.Fatalf("switch exhaust fan to same unit: %v", err)
	}

	target, ok := linkage.Target("OR-1")
	if !ok || target != "EX-1" {
		t.Fatalf("linkage target = %q, want EX-1", target)
	}
	if got := controller.FanState("EX-1"); got != model.FanRunning {
		t.Fatalf("EX-1 state = %v, want running", got)
	}
}
