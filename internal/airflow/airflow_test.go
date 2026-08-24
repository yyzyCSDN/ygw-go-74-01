package airflow

import (
	"context"
	"errors"
	"testing"

	"cleanroomorcontrol/internal/model"
)

// stubDriver lets tests decide which fan fails to start or stop.
type stubDriver struct {
	startErr map[model.FanID]error
	stopErr  map[model.FanID]error
	started  map[model.FanID]bool
	stopped  map[model.FanID]bool
}

func newStubDriver() *stubDriver {
	return &stubDriver{
		startErr: make(map[model.FanID]error),
		stopErr:  make(map[model.FanID]error),
		started:  make(map[model.FanID]bool),
		stopped:  make(map[model.FanID]bool),
	}
}

func (d *stubDriver) Start(fan model.FanID) error {
	if err, ok := d.startErr[fan]; ok {
		return err
	}
	d.started[fan] = true
	return nil
}

func (d *stubDriver) Stop(fan model.FanID) error {
	if err, ok := d.stopErr[fan]; ok {
		return err
	}
	d.stopped[fan] = true
	return nil
}

func newSupplyController(driver FanDriver) *Controller {
	units := []model.FanUnit{
		{ID: "SUP-1", Role: model.FanSupply, State: model.FanRunning, Airflow: 4200},
		{ID: "SUP-2", Role: model.FanSupply, State: model.FanStandby, Airflow: 4100},
	}
	c := NewController(units, driver, nil, nil)
	c.RegisterRoom("OR-1")
	return c
}

// TestSwitchSupplyFan_StartFailureSurfacesError reproduces the original bug:
// when the backup supply fan fails to start, the switch must NOT be reported as
// successful. The error has to come back, the failed fan marked FanFailed, and
// the attempt recorded as a failure rather than a success event.
func TestSwitchSupplyFan_StartFailureSurfacesError(t *testing.T) {
	driver := newStubDriver()
	driver.startErr["SUP-2"] = ErrFanStartFailed
	c := newSupplyController(driver)

	err := c.SwitchSupplyFan(context.Background(), "OR-1", "SUP-2")
	if err == nil {
		t.Fatal("expected switch to fail when backup supply fan cannot start, got nil")
	}
	var se *SwitchError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SwitchError, got %T", err)
	}
	if se.Op != "start" {
		t.Fatalf("expected op=start, got %q", se.Op)
	}
	if !errors.Is(err, ErrFanStartFailed) {
		t.Fatalf("expected wrapped ErrFanStartFailed, got %v", err)
	}

	// The failed fan must be marked failed, not running, or the console would
	// still believe supply is up.
	if got := c.FanState("SUP-2"); got != model.FanFailed {
		t.Fatalf("failed supply fan state = %s, want failed", got)
	}
	// It must be recorded as a failure, not a success event.
	if len(c.FailureEvents()) != 1 {
		t.Fatalf("failure events = %d, want 1", len(c.FailureEvents()))
	}
	if len(c.Events()) != 0 {
		t.Fatalf("success events = %d, want 0", len(c.Events()))
	}
}

// TestSwitchSupplyFan_RetryAfterFailure ensures the operator can retry the
// switch after a failed start: once the underlying fault clears, a second
// attempt must bring the backup online and report success.
func TestSwitchSupplyFan_RetryAfterFailure(t *testing.T) {
	driver := newStubDriver()
	driver.startErr["SUP-2"] = ErrFanStartFailed
	c := newSupplyController(driver)

	if err := c.SwitchSupplyFan(context.Background(), "OR-1", "SUP-2"); err == nil {
		t.Fatal("first switch attempt should have failed")
	}

	// Fault clears; retry.
	delete(driver.startErr, "SUP-2")
	if err := c.SwitchSupplyFan(context.Background(), "OR-1", "SUP-2"); err != nil {
		t.Fatalf("retry should succeed, got %v", err)
	}
	if got := c.FanState("SUP-2"); got != model.FanRunning {
		t.Fatalf("retried supply fan state = %s, want running", got)
	}
}
