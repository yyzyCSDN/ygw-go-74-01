package airflow

import (
	"testing"

	"cleanroomorcontrol/internal/model"
)

func TestSummarizeModes(t *testing.T) {
	modes := map[model.RoomID]model.AirflowMode{
		"OR-1": model.ModeNormal,
		"CR-1": model.ModePurge,
	}
	summary := SummarizeModes(modes)
	if summary["OR-1"] != "normal" || summary["CR-1"] != "purge" {
		t.Fatalf("summary = %v", summary)
	}
}

func TestModeStable(t *testing.T) {
	state := model.AirflowState{Mode: model.ModeNormal, Stable: true}
	if !ModeStable(state) {
		t.Fatal("normal stable must report stable")
	}
	state.Stable = false
	if ModeStable(state) {
		t.Fatal("unstable state must not report stable")
	}
}

func TestHealthyFans(t *testing.T) {
	units := []model.FanUnit{
		{ID: "SUP-1", Role: model.FanSupply, State: model.FanRunning},
		{ID: "SUP-2", Role: model.FanSupply, State: model.FanStandby},
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning},
	}
	fans := HealthyFans(units, model.FanSupply)
	if len(fans) != 1 || fans[0] != "SUP-1" {
		t.Fatalf("healthy supply fans = %v", fans)
	}
	active := ActiveRoleFan(units, model.FanExhaust)
	if active != "EX-1" {
		t.Fatalf("active exhaust = %v", active)
	}
}
