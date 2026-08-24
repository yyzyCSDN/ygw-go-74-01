package alarm

import (
	"testing"
	"time"

	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

// mutableSource is a PointSource whose underlying table can be updated between
// cache reads, mirroring how Monitor.WriteSample mutates the point table.
type mutableSource struct {
	table *particle.Table
}

func (s *mutableSource) Read(point model.PointID) (particle.Point, bool) {
	return s.table.Read(point)
}

// TestCacheGetReloadsAfterInvalidate proves the core of the bug report: once a
// point has been cached, a fresh write to the point table must become visible to
// the next Get only after Invalidate drops the cached copy. Without the
// invalidation, Get keeps serving the stale (pre-over-limit) reading and the
// alarm never fires on the cycle that crossed the limit.
func TestCacheGetReloadsAfterInvalidate(t *testing.T) {
	table := particle.NewTable()
	table.Register("P-01", "OR-1", 100)
	table.Write("P-01", 10, 1.0, time.Now())

	cache := NewCache(&mutableSource{table: table})

	// Prime the cache with the in-limit reading.
	first, ok := cache.Get("P-01")
	if !ok || first.Count != 10 || first.HasData != true {
		t.Fatalf("prime reading = %+v ok=%v", first, ok)
	}

	// A new sample crosses the limit and is written to the point table.
	table.Write("P-01", 150, 1.0, time.Now())

	// Without invalidation the cache still serves the stale in-limit value...
	stale, _ := cache.Get("P-01")
	if stale.Count != 10 {
		t.Fatalf("expected stale cached count 10 before invalidation, got %d", stale.Count)
	}

	// ...but after invalidation the next Get must reload the fresh over-limit
	// reading from the source so the alarm center judges against it.
	cache.Invalidate("P-01")
	fresh, ok := cache.Get("P-01")
	if !ok || fresh.Count != 150 {
		t.Fatalf("expected reloaded count 150 after invalidation, got %d ok=%v", fresh.Count, ok)
	}
	if !fresh.OverLimit() {
		t.Fatalf("reloaded reading must be over limit: %+v", fresh)
	}
}

// TestCacheInvalidateDropsEntry guards against the old broken semantics, which
// left a stale entry (Count = -1) in the map. Get serves any present entry, so
// such a "stale marker" would short-circuit the reload and instead feed the alarm
// center an invalid reading that suppresses the alert.
func TestCacheInvalidateDropsEntry(t *testing.T) {
	table := particle.NewTable()
	table.Register("P-02", "CR-1", 100)
	table.Write("P-02", 5, 1.0, time.Now())

	cache := NewCache(&mutableSource{table: table})
	if _, ok := cache.Get("P-02"); !ok {
		t.Fatal("expected cache to load point on first Get")
	}
	if cache.Size() != 1 {
		t.Fatalf("expected size 1 after load, got %d", cache.Size())
	}

	cache.Invalidate("P-02")
	if cache.Size() != 0 {
		t.Fatalf("expected size 0 after invalidation, got %d", cache.Size())
	}
}
