package service

import (
	"context"
	"path/filepath"

	"cleanroomorcontrol/internal/airflow"
	"cleanroomorcontrol/internal/alarm"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/disinfection"
	"cleanroomorcontrol/internal/door"
	"cleanroomorcontrol/internal/env"
	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type Config struct {
	DataDir string
}

type SettleSignals struct {
	settle map[model.RoomID]chan struct{}
}

func NewSettleSignals() *SettleSignals {
	return &SettleSignals{settle: make(map[model.RoomID]chan struct{})}
}

func (s *SettleSignals) SettleChannel(room model.RoomID) <-chan struct{} {
	ch, ok := s.settle[room]
	if !ok {
		ch = make(chan struct{}, 1)
		s.settle[room] = ch
	}
	return ch
}

func (s *SettleSignals) MarkSettled(room model.RoomID) {
	ch, ok := s.settle[room]
	if !ok {
		ch = make(chan struct{}, 1)
		s.settle[room] = ch
	}
	select {
	case ch <- struct{}{}:
	default:
	}
}

type SimFanDriver struct {
	failSupply map[model.FanID]bool
}

func NewSimFanDriver() *SimFanDriver {
	return &SimFanDriver{failSupply: make(map[model.FanID]bool)}
}

func (d *SimFanDriver) FailSupply(fan model.FanID) {
	d.failSupply[fan] = true
}

func (d *SimFanDriver) Start(fan model.FanID) error {
	if d.failSupply[fan] {
		return airflow.ErrFanStartFailed
	}
	return nil
}

func (d *SimFanDriver) Stop(fan model.FanID) error {
	return nil
}

type airflowVentRunner struct {
	controller *airflow.Controller
}

func (a airflowVentRunner) Run(ctx context.Context, room model.RoomID) error {
	return a.controller.Ventilate(ctx, room)
}

func DefaultRooms() []model.Room {
	return []model.Room{
		{
			ID:               "OR-1",
			Name:             "Operating Room 1",
			CleanClass:       5,
			PressureTarget:   25,
			PressureLow:      15,
			PressureCritical: 5,
			ParticleLimit:    100,
			TempMin:          20,
			TempMax:          24,
			HumidityMin:      40,
			HumidityMax:      60,
		},
		{
			ID:               "CR-1",
			Name:             "Clean Room 1",
			CleanClass:       6,
			PressureTarget:   20,
			PressureLow:      12,
			PressureCritical: 4,
			ParticleLimit:    200,
			TempMin:          18,
			TempMax:          26,
			HumidityMin:      35,
			HumidityMax:      65,
		},
	}
}

func DefaultFans() []model.FanUnit {
	return []model.FanUnit{
		{ID: "SUP-1", Role: model.FanSupply, State: model.FanRunning, Airflow: 4200},
		{ID: "SUP-2", Role: model.FanSupply, State: model.FanStandby, Airflow: 4100},
		{ID: "EX-1", Role: model.FanExhaust, State: model.FanRunning, Airflow: 3600},
		{ID: "EX-2", Role: model.FanExhaust, State: model.FanStandby, Airflow: 3550},
	}
}

func BuildService(cfg Config) (*Service, error) {
	rooms := DefaultRooms()
	fans := DefaultFans()
	linkage := differential.NewLinkage()
	for _, room := range rooms {
		active := airflow.ActiveRoleFan(fans, model.FanExhaust)
		if active != "" {
			linkage.SetTarget(room.ID, active)
		}
	}
	driver := NewSimFanDriver()
	signals := NewSettleSignals()
	controller := airflow.NewController(fans, driver, linkage, signals)
	for _, room := range rooms {
		controller.RegisterRoom(room.ID)
	}
	table := particle.NewTable()
	for _, room := range rooms {
		table.Register(model.PointID(room.ID+"-P1"), room.ID, room.ParticleLimit)
		table.Register(model.PointID(room.ID+"-P2"), room.ID, room.ParticleLimit)
	}
	alarmCache := alarm.NewCache(table)
	center := alarm.NewCenter(alarmCache, alarm.NewLogNotifier())
	pressure := differential.NewMonitor(rooms, linkage, center, controller)
	writer := particle.NewRecordWriter(filepath.Join(cfg.DataDir, "records"), nil)
	particleMonitor := particle.NewMonitor(table, writer, alarmCache)
	doorController := door.NewController(signals)
	for _, room := range rooms {
		doorController.RegisterDoor(string(room.ID)+"-DOOR", room.ID)
	}
	planner := disinfection.NewPlanner(doorController, airflowVentRunner{controller: controller})
	for _, room := range rooms {
		planner.RegisterRoom(room.ID)
	}
	limits := env.NewLimitsStore()
	for _, room := range rooms {
		limits.Set(room.ID, model.EnvLimits{
			TempMin:     room.TempMin,
			TempMax:     room.TempMax,
			HumidityMin: room.HumidityMin,
			HumidityMax: room.HumidityMax,
		})
	}
	envMonitor := env.NewMonitor(limits, center)
	store := NewFileSnapshotStore(filepath.Join(cfg.DataDir, "meta"))
	service := &Service{
		Rooms:         rooms,
		Differential:  pressure,
		Airflow:       controller,
		Particle:      particleMonitor,
		Door:          doorController,
		Disinfection:  planner,
		Env:           envMonitor,
		Alarm:         center,
		SnapshotStore: store,
		Signals:       signals,
		Driver:        driver,
	}
	return service, nil
}
