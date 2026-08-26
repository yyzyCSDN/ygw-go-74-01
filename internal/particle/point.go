package particle

import (
	"sync"
	"time"

	"cleanroomorcontrol/internal/model"
)

type Point struct {
	ID         model.PointID
	Room       model.RoomID
	Limit      int
	LastCount  int
	LastVolume float64
	HasSample  bool
	UpdatedAt  time.Time
}

type Table struct {
	mu      sync.RWMutex
	entries map[model.PointID]Point
}

func NewTable() *Table {
	return &Table{entries: make(map[model.PointID]Point)}
}

func (t *Table) Register(point model.PointID, room model.RoomID, limit int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.entries[point]
	if !ok {
		entry = Point{ID: point, Room: room, Limit: limit}
	}
	entry.Limit = limit
	t.entries[point] = entry
}

func (t *Table) Write(point model.PointID, count int, volume float64, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.entries[point]
	if !ok {
		return
	}
	entry.LastCount = count
	entry.LastVolume = volume
	entry.HasSample = true
	entry.UpdatedAt = at
	t.entries[point] = entry
}

func (t *Table) Read(point model.PointID) (Point, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	entry, ok := t.entries[point]
	return entry, ok
}

func (t *Table) Points(room model.RoomID) []Point {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Point
	for _, entry := range t.entries {
		if entry.Room == room {
			out = append(out, entry)
		}
	}
	return out
}

func (t *Table) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.entries)
}
