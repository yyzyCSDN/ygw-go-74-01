package model

import (
	"fmt"
	"time"
)

type RoomID string

type PointID string

type FanID string

type Room struct {
	ID               RoomID
	Name             string
	CleanClass       int
	PressureTarget   float64
	PressureLow      float64
	PressureCritical float64
	ParticleLimit    int
	TempMin          float64
	TempMax          float64
	HumidityMin      float64
	HumidityMax      float64
}

type PressureStatus int

const (
	PressureStable PressureStatus = iota
	PressureDrooping
	PressureAlarm
	PressureRestoring
)

func (s PressureStatus) String() string {
	switch s {
	case PressureStable:
		return "stable"
	case PressureDrooping:
		return "drooping"
	case PressureAlarm:
		return "alarm"
	case PressureRestoring:
		return "restoring"
	default:
		return fmt.Sprintf("pressure(%d)", int(s))
	}
}

type AirflowMode int

const (
	ModeNormal AirflowMode = iota
	ModePurge
	ModeStandby
)

func (m AirflowMode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModePurge:
		return "purge"
	case ModeStandby:
		return "standby"
	default:
		return fmt.Sprintf("mode(%d)", int(m))
	}
}

func CanTransitMode(from, to AirflowMode) bool {
	switch from {
	case ModeNormal:
		return to == ModePurge
	case ModePurge:
		return to == ModeStandby
	case ModeStandby:
		return to == ModeNormal
	default:
		return false
	}
}

type FanState int

const (
	FanRunning FanState = iota
	FanStandby
	FanFailed
	FanStopped
)

func (s FanState) String() string {
	switch s {
	case FanRunning:
		return "running"
	case FanStandby:
		return "standby"
	case FanFailed:
		return "failed"
	case FanStopped:
		return "stopped"
	default:
		return fmt.Sprintf("fan(%d)", int(s))
	}
}

type FanRole int

const (
	FanSupply FanRole = iota
	FanExhaust
)

func (r FanRole) String() string {
	if r == FanSupply {
		return "supply"
	}
	return "exhaust"
}

type DisinfectionPhase int

const (
	PhaseIdle DisinfectionPhase = iota
	PhaseDisinfecting
	PhaseVentilating
)

func (p DisinfectionPhase) String() string {
	switch p {
	case PhaseIdle:
		return "idle"
	case PhaseDisinfecting:
		return "disinfecting"
	case PhaseVentilating:
		return "ventilating"
	default:
		return fmt.Sprintf("phase(%d)", int(p))
	}
}

func CanTransitPhase(from, to DisinfectionPhase) bool {
	switch from {
	case PhaseIdle:
		return to == PhaseDisinfecting
	case PhaseDisinfecting:
		return to == PhaseVentilating
	case PhaseVentilating:
		return to == PhaseIdle
	default:
		return false
	}
}

type DoorState int

const (
	DoorUnlocked DoorState = iota
	DoorLocked
	DoorInterlocked
)

func (s DoorState) String() string {
	switch s {
	case DoorUnlocked:
		return "unlocked"
	case DoorLocked:
		return "locked"
	case DoorInterlocked:
		return "interlocked"
	default:
		return fmt.Sprintf("door(%d)", int(s))
	}
}

func CanEnterDoor(state DoorState) bool {
	return state == DoorUnlocked
}

type AlarmRecord struct {
	ID      string
	Room    RoomID
	Kind    string
	Message string
	Level   string
	Active  bool
	At      time.Time
}

func (a AlarmRecord) Key() string {
	return a.ID
}
