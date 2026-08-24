package model

import (
	"fmt"
	"time"
)

type FanUnit struct {
	ID      FanID
	Role    FanRole
	State   FanState
	Airflow float64
}

type AirflowState struct {
	Mode   AirflowMode
	Stable bool
	Since  time.Time
}

func (a AirflowState) String() string {
	return fmt.Sprintf("%s stable=%v", a.Mode, a.Stable)
}

type SwitchEvent struct {
	Room RoomID
	From FanID
	To   FanID
	At   time.Time
}

func (e SwitchEvent) Summary() string {
	return fmt.Sprintf("room=%s from=%s to=%s at=%s", e.Room, e.From, e.To, e.At.Format("15:04:05"))
}
