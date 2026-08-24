package alarm

import (
	"sync"

	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type PointSource interface {
	Read(point model.PointID) (particle.Point, bool)
}

type Cache struct {
	mu          sync.RWMutex
	source      PointSource
	entries     map[model.PointID]model.ParticleReading
	buckets     map[uint64][]model.PointID
	epoch       int
	bucketCount int
}

func NewCache(source PointSource) *Cache {
	return &Cache{
		source:      source,
		entries:     make(map[model.PointID]model.ParticleReading),
		buckets:     make(map[uint64][]model.PointID),
		bucketCount: 8,
	}
}

func (c *Cache) bucket(point model.PointID) uint64 {
	return particle.PointHash(point) % uint64(c.bucketCount)
}

func (c *Cache) Get(point model.PointID) (model.ParticleReading, bool) {
	c.mu.RLock()
	entry, ok := c.entries[point]
	c.mu.RUnlock()
	if ok {
		return entry, true
	}
	return c.load(point)
}

func (c *Cache) load(point model.PointID) (model.ParticleReading, bool) {
	source, ok := c.source.Read(point)
	if !ok {
		return model.ParticleReading{}, false
	}
	reading := model.ParticleReading{
		Point:   point,
		Room:    source.Room,
		Count:   source.LastCount,
		Volume:  source.LastVolume,
		At:      source.UpdatedAt,
		Limit:   source.Limit,
		HasData: source.HasSample,
	}
	c.mu.Lock()
	c.entries[point] = reading
	bucket := c.bucket(point)
	found := false
	for _, member := range c.buckets[bucket] {
		if member == point {
			found = true
			break
		}
	}
	if !found {
		c.buckets[bucket] = append(c.buckets[bucket], point)
	}
	c.mu.Unlock()
	return reading, true
}

// Invalidate drops the cached reading for point so the next Get reloads the
// current value from the source. Marking the entry stale in place would leave a
// stale ParticleReading in the map, and Get serves any present entry directly —
// so removing the entry is what actually forces a reload.
func (c *Cache) Invalidate(point model.PointID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, point)
	bucket := c.bucket(point)
	members := c.buckets[bucket]
	for i, member := range members {
		if member == point {
			c.buckets[bucket] = append(members[:i], members[i+1:]...)
			break
		}
	}
	c.epoch++
}

func (c *Cache) InvalidateBucket(point model.PointID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	bucket := c.bucket(point)
	for _, member := range c.buckets[bucket] {
		delete(c.entries, member)
	}
	delete(c.buckets, bucket)
	c.epoch++
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
