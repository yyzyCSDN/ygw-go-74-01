package airflow

import (
	"context"
	"sync"
	"time"

	"cleanroomorcontrol/internal/model"
)

type LinkageBroker interface {
	ApplyFanSwitch(room model.RoomID, fan model.FanID)
	Target(room model.RoomID) (model.FanID, bool)
}

type Controller struct {
	mu       sync.RWMutex
	units    map[model.FanID]model.FanUnit
	driver   FanDriver
	modes    map[model.RoomID]model.AirflowMode
	states   map[model.RoomID]model.AirflowState
	linkage  LinkageBroker
	signals  SignalSource
	events   []model.SwitchEvent
	failures []model.SwitchEvent
}

func NewController(units []model.FanUnit, driver FanDriver, linkage LinkageBroker, signals SignalSource) *Controller {
	controller := &Controller{
		units:   make(map[model.FanID]model.FanUnit),
		driver:  driver,
		modes:   make(map[model.RoomID]model.AirflowMode),
		states:  make(map[model.RoomID]model.AirflowState),
		linkage: linkage,
		signals: signals,
	}
	for _, unit := range units {
		controller.units[unit.ID] = unit
	}
	return controller
}

func (c *Controller) RegisterRoom(room model.RoomID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modes[room] = model.ModeNormal
	c.states[room] = model.AirflowState{Mode: model.ModeNormal, Stable: true, Since: time.Now()}
}

func (c *Controller) FanState(fan model.FanID) model.FanState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	unit, ok := c.units[fan]
	if !ok {
		return model.FanStopped
	}
	return unit.State
}

func (c *Controller) Mode(room model.RoomID) model.AirflowMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modes[room]
}

func (c *Controller) Stable(room model.RoomID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.states[room].Stable
}

func (c *Controller) Events() []model.SwitchEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.SwitchEvent, len(c.events))
	copy(out, c.events)
	return out
}

func (c *Controller) FailureEvents() []model.SwitchEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.SwitchEvent, len(c.failures))
	copy(out, c.failures)
	return out
}

func (c *Controller) Units() []model.FanUnit {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.FanUnit, 0, len(c.units))
	for _, unit := range c.units {
		out = append(out, unit)
	}
	return out
}

func (c *Controller) SwitchExhaustFan(ctx context.Context, room model.RoomID, target model.FanID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.switchExhaustFanLocked(room, target)
}

func (c *Controller) switchExhaustFanLocked(room model.RoomID, target model.FanID) error {
	unit, ok := c.units[target]
	if !ok {
		return &SwitchError{Room: room, Fan: target, Op: "switch", Err: ErrUnknownFan}
	}
	if unit.Role != model.FanExhaust {
		return &SwitchError{Room: room, Fan: target, Op: "switch", Err: ErrInvalidModeTransition}
	}
	old := c.currentExhaust(room)
	if err := c.driver.Stop(old); err != nil {
		return &SwitchError{Room: room, Fan: target, Op: "stop", Err: err}
	}
	if err := c.driver.Start(target); err != nil {
		c.units[target] = model.FanUnit{ID: target, Role: unit.Role, State: model.FanFailed, Airflow: unit.Airflow}
		c.failures = append(c.failures, model.SwitchEvent{Room: room, From: old, To: target, At: time.Now()})
		return &SwitchError{Room: room, Fan: target, Op: "start", Err: err}
	}
	c.units[target] = model.FanUnit{ID: target, Role: unit.Role, State: model.FanRunning, Airflow: unit.Airflow}
	c.refreshLinkage(room, target)
	return nil
}

func (c *Controller) currentExhaust(room model.RoomID) model.FanID {
	if c.linkage != nil {
		target, ok := c.linkage.Target(room)
		if ok && c.units[target].Role == model.FanExhaust {
			return target
		}
	}
	for id, unit := range c.units {
		if unit.Role == model.FanExhaust && unit.State == model.FanRunning {
			return id
		}
	}
	return ""
}

func (c *Controller) refreshLinkage(room model.RoomID, target model.FanID) {
	if c.linkage == nil {
		return
	}
	previous := c.currentExhaust(room)
	if previous != "" && previous == target {
		return
	}
	if !c.targetUsable(room, target) {
		return
	}
	c.linkage.ApplyFanSwitch(room, target)
	if confirmed, ok := c.linkage.Target(room); !ok || confirmed != target {
		return
	}
	if previous != "" && previous != target {
		c.events = append(c.events, model.SwitchEvent{Room: room, From: previous, To: target, At: time.Now()})
	}
}

func (c *Controller) targetUsable(room model.RoomID, target model.FanID) bool {
	unit, ok := c.units[target]
	if !ok {
		return false
	}
	if unit.State == model.FanFailed {
		return false
	}
	return unit.Role == model.FanExhaust
}

func (c *Controller) SwitchSupplyFan(ctx context.Context, room model.RoomID, target model.FanID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.switchSupplyFanLocked(room, target)
}

func (c *Controller) switchSupplyFanLocked(room model.RoomID, target model.FanID) error {
	unit, ok := c.units[target]
	if !ok {
		return &SwitchError{Room: room, Fan: target, Op: "switch", Err: ErrUnknownFan}
	}
	if unit.Role != model.FanSupply {
		return &SwitchError{Room: room, Fan: target, Op: "switch", Err: ErrInvalidModeTransition}
	}
	old := c.currentSupply(room)
	if err := c.driver.Stop(old); err != nil {
		return &SwitchError{Room: room, Fan: target, Op: "stop", Err: err}
	}
	if err := c.driver.Start(target); err != nil {
		c.units[target] = model.FanUnit{ID: target, Role: unit.Role, State: model.FanFailed, Airflow: unit.Airflow}
		c.failures = append(c.failures, model.SwitchEvent{Room: room, From: old, To: target, At: time.Now()})
		return &SwitchError{Room: room, Fan: target, Op: "start", Err: err}
	}
	c.units[target] = model.FanUnit{ID: target, Role: unit.Role, State: model.FanRunning, Airflow: unit.Airflow}
	c.recordSupplySwitch(room, old, target)
	return nil
}

func (c *Controller) recordSupplySwitch(room model.RoomID, old model.FanID, target model.FanID) {
	c.events = append(c.events, model.SwitchEvent{Room: room, From: old, To: target, At: time.Now()})
}

func (c *Controller) currentSupply(room model.RoomID) model.FanID {
	for id, unit := range c.units {
		if unit.Role == model.FanSupply && unit.State == model.FanRunning {
			return id
		}
	}
	return ""
}

func (c *Controller) SwitchAirflow(ctx context.Context, room model.RoomID, mode model.AirflowMode) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.modes[room]
	if !model.CanTransitMode(current, mode) {
		return &SwitchError{Room: room, Op: "mode", Err: ErrInvalidModeTransition}
	}
	if err := WaitStable(ctx, room, c.signals); err != nil {
		c.states[room] = model.AirflowState{Mode: current, Stable: false, Since: time.Now()}
		return &SwitchError{Room: room, Op: "wait", Err: err}
	}
	c.modes[room] = mode
	c.states[room] = model.AirflowState{Mode: mode, Stable: true, Since: time.Now()}
	if c.modes[room] != mode {
		return &SwitchError{Room: room, Op: "apply", Err: ErrInvalidModeTransition}
	}
	if c.states[room].Mode != mode {
		return &SwitchError{Room: room, Op: "apply", Err: ErrInvalidModeTransition}
	}
	return nil
}

func (c *Controller) Purge(ctx context.Context, room model.RoomID) error {
	return c.SwitchAirflow(ctx, room, model.ModePurge)
}

func (c *Controller) Standby(ctx context.Context, room model.RoomID) error {
	return c.SwitchAirflow(ctx, room, model.ModeStandby)
}

func (c *Controller) Ventilate(ctx context.Context, room model.RoomID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.modes[room]
	if current == model.ModePurge {
		return nil
	}
	if !model.CanTransitMode(current, model.ModePurge) {
		return &SwitchError{Room: room, Op: "ventilate", Err: ErrInvalidModeTransition}
	}
	_ = WaitStable(ctx, room, c.signals)
	c.modes[room] = model.ModePurge
	c.states[room] = model.AirflowState{Mode: model.ModePurge, Stable: true, Since: time.Now()}
	return nil
}
