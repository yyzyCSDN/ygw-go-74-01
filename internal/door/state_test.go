package door

import (
	"testing"

	"cleanroomorcontrol/internal/model"
)

func TestSummarizeDoors(t *testing.T) {
	states := map[model.RoomID]model.DoorState{
		"OR-1": model.DoorUnlocked,
		"CR-1": model.DoorInterlocked,
	}
	summary := SummarizeDoors(states)
	if summary["OR-1"] != "unlocked" || summary["CR-1"] != "interlocked" {
		t.Fatalf("summary = %v", summary)
	}
}

func TestAllowedEntry(t *testing.T) {
	states := map[model.RoomID]model.DoorState{
		"OR-1": model.DoorUnlocked,
		"CR-1": model.DoorLocked,
		"OR-2": model.DoorInterlocked,
	}
	allowed := AllowedEntry(states)
	if len(allowed) != 1 || allowed[0] != "OR-1" {
		t.Fatalf("allowed = %v", allowed)
	}
}
