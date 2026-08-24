package airflow_test

import (
	"context"
	"errors"
	"testing"

	"cleanroomorcontrol/internal/airflow"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

type failingProbeDriver struct{}

func (failingProbeDriver) Start(fan model.FanID) error { return errors.New("motor fault") }

func (failingProbeDriver) Stop(fan model.FanID) error { return nil }

type probeSignalSource struct{}

func (probeSignalSource) SettleChannel(room model.RoomID) <-chan struct{} {
	return make(chan struct{})
}

func TestFanSwitchErrorNotSwallowed(t *testing.T) {
	linkage := differential.NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	units := []model.FanUnit{
		{ID: "SUP-1", Role: model.FanSupply, State: model.FanRunning},
		{ID: "SUP-2", Role: model.FanSupply, State: model.FanStandby},
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning},
	}
	controller := airflow.NewController(units, failingProbeDriver{}, linkage, probeSignalSource{})
	controller.RegisterRoom("OR-1")
	err := controller.SwitchSupplyFan(context.Background(), "OR-1", "SUP-2")
	if err == nil {
		t.Fatal("switch failure must be reported")
	}
	if controller.FanState("SUP-2") == model.FanRunning {
		t.Fatal("failed switch must not mark fan running")
	}
}
