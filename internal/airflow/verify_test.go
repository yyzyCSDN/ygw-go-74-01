package airflow_test

import (
	"context"
	"testing"
	"time"

	"cleanroomorcontrol/internal/airflow"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/model"
)

type okProbeDriver struct{}

func (okProbeDriver) Start(fan model.FanID) error { return nil }

func (okProbeDriver) Stop(fan model.FanID) error { return nil }

type idleProbeSource struct{}

func (idleProbeSource) SettleChannel(room model.RoomID) <-chan struct{} {
	return make(chan struct{})
}

func TestAirflowWaitTimeoutHandled(t *testing.T) {
	linkage := differential.NewLinkage()
	linkage.SetTarget("OR-1", "EX-1")
	units := []model.FanUnit{
		{ID: "SUP-1", Role: model.FanSupply, State: model.FanRunning},
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning},
	}
	controller := airflow.NewController(units, okProbeDriver{}, linkage, idleProbeSource{})
	controller.RegisterRoom("OR-1")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	err := controller.SwitchAirflow(ctx, "OR-1", model.ModePurge)
	if err == nil {
		t.Fatal("wait timeout must fail the switch")
	}
	if controller.Mode("OR-1") == model.ModePurge {
		t.Fatal("switch must not be marked success after timeout")
	}
	if controller.Stable("OR-1") {
		t.Fatal("room must not be stable after timeout")
	}
}
