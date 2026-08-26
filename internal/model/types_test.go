package model

import (
	"fmt"
	"testing"
)

func TestStatusStrings(t *testing.T) {
	cases := []struct {
		value fmt.Stringer
		want  string
	}{
		{PressureStable, "stable"},
		{PressureDrooping, "drooping"},
		{PressureAlarm, "alarm"},
		{PressureRestoring, "restoring"},
		{ModeNormal, "normal"},
		{ModePurge, "purge"},
		{ModeStandby, "standby"},
		{FanRunning, "running"},
		{FanFailed, "failed"},
		{PhaseIdle, "idle"},
		{PhaseDisinfecting, "disinfecting"},
		{PhaseVentilating, "ventilating"},
		{DoorUnlocked, "unlocked"},
		{DoorInterlocked, "interlocked"},
	}
	for _, item := range cases {
		if got := item.value.String(); got != item.want {
			t.Fatalf("String() = %q, want %q", got, item.want)
		}
	}
}

func TestModeTransitions(t *testing.T) {
	if !CanTransitMode(ModeNormal, ModePurge) {
		t.Fatal("normal to purge must be allowed")
	}
	if !CanTransitMode(ModePurge, ModeStandby) {
		t.Fatal("purge to standby must be allowed")
	}
	if !CanTransitMode(ModeStandby, ModeNormal) {
		t.Fatal("standby to normal must be allowed")
	}
	if CanTransitMode(ModeNormal, ModeStandby) {
		t.Fatal("normal to standby must be rejected")
	}
}

func TestPhaseTransitions(t *testing.T) {
	if !CanTransitPhase(PhaseIdle, PhaseDisinfecting) {
		t.Fatal("idle to disinfecting must be allowed")
	}
	if !CanTransitPhase(PhaseDisinfecting, PhaseVentilating) {
		t.Fatal("disinfecting to ventilating must be allowed")
	}
	if !CanTransitPhase(PhaseVentilating, PhaseIdle) {
		t.Fatal("ventilating to idle must be allowed")
	}
	if CanTransitPhase(PhaseIdle, PhaseVentilating) {
		t.Fatal("idle to ventilating must be rejected")
	}
}

func TestDoorEntry(t *testing.T) {
	if !CanEnterDoor(DoorUnlocked) {
		t.Fatal("unlocked door must allow entry")
	}
	if CanEnterDoor(DoorLocked) || CanEnterDoor(DoorInterlocked) {
		t.Fatal("locked or interlocked door must deny entry")
	}
}

func TestAlarmRecordKey(t *testing.T) {
	record := AlarmRecord{ID: "pressure:OR-1"}
	if record.Key() != "pressure:OR-1" {
		t.Fatal("alarm record key mismatch")
	}
}
