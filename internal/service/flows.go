package service

import (
	"context"
	"errors"

	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

func (s *Service) HandlePressureSample(reading model.PressureReading) model.PressureStatus {
	status := s.Differential.Sample(reading)
	if status == model.PressureDrooping || status == model.PressureAlarm {
		target := s.Differential.MaintenanceTarget(reading.Room)
		if target == "" {
			s.Alarm.ReportPressure(reading.Room, model.PressureAlarm, reading.Pa)
		}
	}
	return status
}

func (s *Service) HandleFanSwitch(ctx context.Context, room model.RoomID, role model.FanRole, target model.FanID) error {
	switch role {
	case model.FanExhaust:
		return s.Airflow.SwitchExhaustFan(ctx, room, target)
	case model.FanSupply:
		return s.Airflow.SwitchSupplyFan(ctx, room, target)
	default:
		return errors.New("unknown fan role")
	}
}

func (s *Service) HandleParticleSample(sample model.ParticleSample) error {
	if err := s.Particle.WriteSample(sample); err != nil {
		return err
	}
	readings := s.Particle.LatestReadings(sample.Room)
	s.Alarm.EvaluateRoom(sample.Room, readings)
	return nil
}

func (s *Service) RunParticleCycle(samples []model.ParticleSample) particle.CycleResult {
	return s.Particle.RunCycle(samples)
}

func (s *Service) HandleDisinfectionStart(room model.RoomID) error {
	return s.Disinfection.Start(room)
}

func (s *Service) HandleDisinfectionComplete(room model.RoomID) error {
	return s.Disinfection.CompleteDisinfection(room)
}

func (s *Service) HandleVentilation(ctx context.Context, room model.RoomID) error {
	return s.Disinfection.Ventilate(ctx, room)
}

func (s *Service) HandleEnvSample(reading model.EnvReading) []string {
	return s.Env.Sample(reading)
}

func (s *Service) Recover() error {
	snapshot, err := s.Differential.Recover(s.SnapshotStore, s.Alarm)
	if err != nil {
		return err
	}
	return s.SnapshotStore.SavePressure(snapshot)
}

func (s *Service) Persist() error {
	snapshot := s.Differential.CaptureSnapshot()
	if err := s.SnapshotStore.SavePressure(snapshot); err != nil {
		return err
	}
	rooms := make([]model.RoomID, 0, len(s.Rooms))
	for _, room := range s.Rooms {
		rooms = append(rooms, room.ID)
	}
	return firstError(s.Particle.RotateAll(rooms))
}

func firstError(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
