package disinfection_test

import (
	"context"
	"testing"

	"cleanroomorcontrol/internal/disinfection"
	"cleanroomorcontrol/internal/door"
	"cleanroomorcontrol/internal/model"
)

type settleProbeSourceB struct{}

func (settleProbeSourceB) SettleChannel(room model.RoomID) <-chan struct{} {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return ch
}

type timeoutProbeRunner struct{}

func (timeoutProbeRunner) Run(ctx context.Context, room model.RoomID) error {
	return disinfection.ErrVentilationTimeout
}

func TestDisinfectionTimeoutNotSwallowed(t *testing.T) {
	doorController := door.NewController(settleProbeSourceB{})
	doorController.RegisterDoor("OR-1-DOOR", "OR-1")
	planner := disinfection.NewPlanner(doorController, timeoutProbeRunner{})
	planner.RegisterRoom("OR-1")
	if err := planner.Start("OR-1"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := planner.CompleteDisinfection("OR-1"); err != nil {
		t.Fatalf("complete disinfection failed: %v", err)
	}
	err := planner.Ventilate(context.Background(), "OR-1")
	if err == nil {
		t.Fatal("ventilation timeout must be reported")
	}
	if planner.Phase("OR-1") == model.PhaseIdle {
		t.Fatal("room must not be marked complete after timeout")
	}
}
