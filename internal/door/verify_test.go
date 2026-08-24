package door_test

import (
	"context"
	"testing"
	"time"

	"cleanroomorcontrol/internal/door"
	"cleanroomorcontrol/internal/model"
)

type lockedProbeSource struct{}

func (lockedProbeSource) SettleChannel(room model.RoomID) <-chan struct{} {
	return make(chan struct{})
}

func TestDoorLockTimeoutCancelsWait(t *testing.T) {
	controller := door.NewController(lockedProbeSource{})
	controller.RegisterDoor("OR-1-DOOR", "OR-1")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- controller.Acquire(ctx, "OR-1")
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("timeout must return an error")
		}
	case <-time.After(600 * time.Millisecond):
		t.Fatal("acquire still waiting after timeout")
	}
	if controller.State("OR-1") != model.DoorUnlocked {
		t.Fatalf("interlock must be released after timeout, state = %v", controller.State("OR-1"))
	}
}
