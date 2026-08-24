package alarm

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"cleanroomorcontrol/internal/model"
)

type Notifier interface {
	Notify(record model.AlarmRecord) error
}

type ParticleAlert struct {
	Point  model.PointID
	Room   model.RoomID
	Count  int
	Limit  int
	Active bool
	Reason string
}

type Center struct {
	mu      sync.RWMutex
	cache   *Cache
	notify  Notifier
	records map[string]model.AlarmRecord
}

func NewCenter(cache *Cache, notify Notifier) *Center {
	return &Center{
		cache:   cache,
		notify:  notify,
		records: make(map[string]model.AlarmRecord),
	}
}

func (c *Center) ReportPressure(room model.RoomID, status model.PressureStatus, pa float64) {
	key := "pressure:" + string(room)
	record := model.AlarmRecord{
		ID:      key,
		Room:    room,
		Kind:    "pressure",
		Message: "pressure " + status.String() + " " + formatPa(pa),
		Level:   severityPressure(status),
		Active:  status == model.PressureAlarm,
		At:      time.Now(),
	}
	c.apply(record)
}

func (c *Center) ReportParticle(room model.RoomID, point model.PointID, count int, limit int) {
	key := "particle:" + string(point)
	alert := ParticleAlert{
		Point:  point,
		Room:   room,
		Count:  count,
		Limit:  limit,
		Active: count > limit,
		Reason: "ok",
	}
	if alert.Active {
		alert.Reason = "over-limit"
	}
	record := model.AlarmRecord{
		ID:      key,
		Room:    room,
		Kind:    "particle",
		Message: "particle " + string(point) + " count " + strconv.Itoa(count) + " limit " + strconv.Itoa(limit),
		Level:   severityParticle(alert),
		Active:  alert.Active,
		At:      time.Now(),
	}
	c.apply(record)
}

func (c *Center) ReportEnv(room model.RoomID, reading model.EnvReading, violations []string) {
	key := "env:" + string(room)
	record := model.AlarmRecord{
		ID:      key,
		Room:    room,
		Kind:    "env",
		Message: "env " + strings.Join(violations, ","),
		Level:   severityEnv(violations),
		Active:  len(violations) > 0,
		At:      time.Now(),
	}
	c.apply(record)
}

func (c *Center) EvaluateParticle(reading *model.ParticleReading, point model.PointID) ParticleAlert {
	if reading == nil {
		return ParticleAlert{Point: point, Active: false, Reason: "no-data"}
	}
	if !reading.HasData {
		return ParticleAlert{Point: point, Room: reading.Room, Active: false, Reason: "no-data"}
	}
	if reading.Count < 0 || reading.Limit <= 0 {
		return ParticleAlert{Point: point, Room: reading.Room, Count: reading.Count, Limit: reading.Limit, Active: false, Reason: "invalid-reading"}
	}
	if reading.Volume <= 0 {
		return ParticleAlert{Point: point, Room: reading.Room, Active: false, Reason: "invalid-volume"}
	}
	if reading.At.IsZero() {
		return ParticleAlert{Point: point, Room: reading.Room, Active: false, Reason: "no-timestamp"}
	}
	alert := ParticleAlert{
		Point:  point,
		Room:   reading.Room,
		Count:  reading.Count,
		Limit:  reading.Limit,
		Active: reading.OverLimit(),
		Reason: "ok",
	}
	if alert.Active {
		alert.Reason = "over-limit"
	}
	return alert
}

func (c *Center) EvaluateParticlePoint(room model.RoomID, point model.PointID) ParticleAlert {
	reading, ok := c.cache.Get(point)
	if !ok {
		return ParticleAlert{Point: point, Room: room, Active: false, Reason: "no-data"}
	}
	return c.EvaluateParticle(&reading, point)
}

func (c *Center) EvaluateRoom(room model.RoomID, points []model.ParticleReading) []ParticleAlert {
	var alerts []ParticleAlert
	for _, reading := range points {
		alert := c.EvaluateParticlePoint(room, reading.Point)
		if alert.Active {
			c.ReportParticle(room, reading.Point, alert.Count, alert.Limit)
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

func (c *Center) apply(record model.AlarmRecord) {
	c.mu.Lock()
	previous, existed := c.records[record.ID]
	c.records[record.ID] = record
	c.mu.Unlock()
	if c.notify == nil {
		return
	}
	if !existed || previous.Active != record.Active {
		_ = c.notify.Notify(record)
	}
}

func (c *Center) Records() []model.AlarmRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.AlarmRecord, 0, len(c.records))
	for _, record := range c.records {
		out = append(out, record)
	}
	return out
}

func (c *Center) ActiveAlarms() []model.AlarmRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []model.AlarmRecord
	for _, record := range c.records {
		if record.Active {
			out = append(out, record)
		}
	}
	return out
}

func (c *Center) HasActive(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	record, ok := c.records[key]
	return ok && record.Active
}
