package differential

import (
	"testing"

	"cleanroomorcontrol/internal/model"
)

func TestPressureStateMachineTransitions(t *testing.T) {
	machine := NewPressureStateMachine(15, 5, 3, 2)
	if machine.Status() != model.PressureStable {
		t.Fatal("initial status must be stable")
	}
	machine.Feed(18)
	machine.Feed(14)
	if machine.Status() != model.PressureStable {
		t.Fatal("single low sample must not drop status")
	}
	machine.Feed(14)
	machine.Feed(14)
	if machine.Status() != model.PressureDrooping {
		t.Fatalf("status = %v, want drooping", machine.Status())
	}
	machine.Feed(3)
	if machine.Status() != model.PressureAlarm {
		t.Fatalf("status = %v, want alarm", machine.Status())
	}
	machine.Feed(16)
	machine.Feed(16)
	if machine.Status() != model.PressureRestoring {
		t.Fatalf("status = %v, want restoring", machine.Status())
	}
	machine.Feed(20)
	if machine.Status() != model.PressureStable {
		t.Fatalf("status = %v, want stable", machine.Status())
	}
}

func TestPressureStateMachineReset(t *testing.T) {
	machine := NewPressureStateMachine(15, 5, 1, 1)
	machine.Feed(3)
	machine.Feed(3)
	if machine.Status() != model.PressureAlarm {
		t.Fatalf("status = %v", machine.Status())
	}
	machine.Reset()
	if machine.Status() != model.PressureStable {
		t.Fatalf("reset status = %v", machine.Status())
	}
	if machine.LastPa() != 0 {
		t.Fatalf("last pa = %v", machine.LastPa())
	}
}
