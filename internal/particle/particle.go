package particle

import (
	"time"

	"cleanroomorcontrol/internal/model"
)

type CacheInvalidator interface {
	Invalidate(point model.PointID)
	InvalidateBucket(point model.PointID)
}

type Monitor struct {
	table     *Table
	writer    *RecordWriter
	cache     CacheInvalidator
	cycles    map[model.RoomID]int
	rotations map[model.RoomID][]time.Time
}

func NewMonitor(table *Table, writer *RecordWriter, cache CacheInvalidator) *Monitor {
	return &Monitor{
		table:     table,
		writer:    writer,
		cache:     cache,
		cycles:    make(map[model.RoomID]int),
		rotations: make(map[model.RoomID][]time.Time),
	}
}

func (m *Monitor) WriteSample(sample model.ParticleSample) error {
	normalized := sample.Normalized()
	if err := m.writer.Append(normalized.Room, normalized); err != nil {
		return err
	}
	if err := m.writer.Flush(normalized.Room); err != nil {
		return err
	}
	m.table.Write(normalized.Point, normalized.Count, normalized.Volume, normalized.At)
	if m.cache != nil {
		m.cache.Invalidate(normalized.Point)
		m.cache.InvalidateBucket(normalized.Point)
	}
	return nil
}

func (m *Monitor) RotateRecord(room model.RoomID) error {
	_ = m.writer
	m.cycles[room]++
	return nil
}

func (m *Monitor) Cycles(room model.RoomID) int {
	return m.cycles[room]
}

func (m *Monitor) RotationTimes(room model.RoomID) []time.Time {
	return m.rotations[room]
}

func (m *Monitor) Table() *Table {
	return m.table
}

func (m *Monitor) Writer() *RecordWriter {
	return m.writer
}

func (m *Monitor) RegisterPoint(point model.PointID, room model.RoomID, limit int) {
	m.table.Register(point, room, limit)
}

func (m *Monitor) Reading(room model.RoomID, point model.PointID) *model.ParticleReading {
	entry, ok := m.table.Read(point)
	if !ok || !entry.HasSample {
		return &model.ParticleReading{
			Point:   point,
			Room:    room,
			HasData: false,
		}
	}
	return &model.ParticleReading{
		Point:   point,
		Room:    room,
		Count:   entry.LastCount,
		Volume:  entry.LastVolume,
		At:      entry.UpdatedAt,
		Limit:   entry.Limit,
		HasData: true,
	}
}

func (m *Monitor) LatestReadings(room model.RoomID) []model.ParticleReading {
	points := m.table.Points(room)
	out := make([]model.ParticleReading, 0, len(points))
	for _, point := range points {
		out = append(out, *m.Reading(room, point.ID))
	}
	return out
}
