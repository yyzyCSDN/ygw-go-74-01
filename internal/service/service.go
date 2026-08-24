package service

import (
	"context"
	"time"

	"cleanroomorcontrol/internal/airflow"
	"cleanroomorcontrol/internal/alarm"
	"cleanroomorcontrol/internal/differential"
	"cleanroomorcontrol/internal/disinfection"
	"cleanroomorcontrol/internal/door"
	"cleanroomorcontrol/internal/env"
	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type Service struct {
	Rooms         []model.Room
	Differential  *differential.Monitor
	Airflow       *airflow.Controller
	Particle      *particle.Monitor
	Door          *door.Controller
	Disinfection  *disinfection.Planner
	Env           *env.Monitor
	Alarm         *alarm.Center
	SnapshotStore differential.SnapshotStore
	Signals       *SettleSignals
	Driver        *SimFanDriver
}

func (s *Service) RoomStatus(room model.RoomID) map[string]any {
	target, _ := s.Differential.LinkageTarget(room)
	targetState := "none"
	if unit, ok := airflow.UnitByID(s.Airflow.Units(), target); ok {
		targetState = unit.State.String()
	}
	ventCompleted := ""
	if at, ok := s.Disinfection.CompletedAt(room); ok {
		ventCompleted = at.Format("15:04:05")
	}
	lastVent := ""
	if at, ok := s.Disinfection.LastVent(room); ok {
		lastVent = at.Format("15:04:05")
	}
	nextDue := ""
	if at, ok := s.Disinfection.NextDue(room, time.Now()); ok {
		nextDue = at.Format("15:04:05")
	}
	envTemp := 0.0
	envHumidity := 0.0
	if reading, ok := s.Env.Latest(room); ok {
		envTemp = reading.TempC
		envHumidity = reading.Humidity
	}
	return map[string]any{
		"room":              room,
		"pressure":          s.Differential.Status(room).String(),
		"last_pa":           s.Differential.LastPa(room),
		"airflow_mode":      s.Airflow.Mode(room).String(),
		"airflow_stable":    s.Airflow.Stable(room),
		"mode_stable":       airflow.ModeStable(model.AirflowState{Mode: s.Airflow.Mode(room), Stable: s.Airflow.Stable(room)}),
		"linkage_target":    target,
		"linkage_state":     targetState,
		"linkage_history":   s.Differential.LinkageHistory(room),
		"door":              s.Door.State(room).String(),
		"phase":             s.Disinfection.Phase(room).String(),
		"cycles":            s.Disinfection.Cycles(room),
		"particle_cycles":   s.Particle.Cycles(room),
		"rotations":         len(s.Particle.RotationTimes(room)),
		"vent_completed_at": ventCompleted,
		"last_vent":         lastVent,
		"vent_attempts":     s.Disinfection.VentAttempts(room),
		"next_due":          nextDue,
		"due_count":         len(s.Disinfection.Due(room, time.Now())),
		"can_enter":         s.Door.CanEnter(room),
		"env_temp":          envTemp,
		"env_humidity":      envHumidity,
		"pressure_alarm":    s.Alarm.HasActive("pressure:" + string(room)),
	}
}

func (s *Service) Status() map[string]any {
	rooms := make([]map[string]any, 0, len(s.Rooms))
	for _, room := range s.Rooms {
		rooms = append(rooms, s.RoomStatus(room.ID))
	}
	recordOpen, recordClosed := s.Particle.Writer().Counters()
	doorStates := s.Door.Snapshot()
	return map[string]any{
		"rooms":           rooms,
		"alarms":          len(s.Alarm.ActiveAlarms()),
		"fan_units":       len(s.Airflow.Units()),
		"point_rows":      s.Particle.Table().Count(),
		"switch_events":   s.SwitchEventSummaries(),
		"switch_failures": len(s.Airflow.FailureEvents()),
		"record_open":     recordOpen,
		"record_closed":   recordClosed,
		"airflow_modes":   s.AirflowModes(),
		"doors":           door.SummarizeDoors(doorStates),
		"allowed_entry":   door.AllowedEntry(doorStates),
		"env":             s.EnvSnapshot(),
		"phases":          disinfection.PhaseMap(s.Disinfection.Phases()),
	}
}

func (s *Service) AlarmList() []model.AlarmRecord {
	return s.Alarm.Records()
}

func (s *Service) DoorStates() map[model.RoomID]model.DoorState {
	return s.Door.Snapshot()
}

func (s *Service) AirflowModes() map[model.RoomID]string {
	modes := make(map[model.RoomID]model.AirflowMode, len(s.Rooms))
	for _, room := range s.Rooms {
		modes[room.ID] = s.Airflow.Mode(room.ID)
	}
	return airflow.SummarizeModes(modes)
}

func (s *Service) EnvSnapshot() map[model.RoomID]env.ReadingAggregate {
	all := s.Env.All()
	out := make(map[model.RoomID]env.ReadingAggregate, len(all))
	for room, reading := range all {
		out[room] = env.Aggregate(room, []model.EnvReading{reading})
	}
	return out
}

func (s *Service) DoorAcquire(ctx context.Context, room model.RoomID) error {
	return s.Door.Acquire(ctx, room)
}

func (s *Service) PurgeRoom(ctx context.Context, room model.RoomID) error {
	return s.Airflow.Purge(ctx, room)
}

func (s *Service) StandbyRoom(ctx context.Context, room model.RoomID) error {
	return s.Airflow.Standby(ctx, room)
}

func (s *Service) ScheduleDisinfection(room model.RoomID, at time.Time) {
	s.Disinfection.Schedule(room, at)
}

func (s *Service) RegisterParticlePoint(point model.PointID, room model.RoomID, limit int) {
	s.Particle.RegisterPoint(point, room, limit)
}

func (s *Service) AlarmActive(key string) bool {
	return s.Alarm.HasActive(key)
}

func (s *Service) SwitchEventSummaries() []string {
	events := s.Airflow.Events()
	out := make([]string, 0, len(events))
	for _, event := range events {
		out = append(out, event.Summary())
	}
	return out
}
