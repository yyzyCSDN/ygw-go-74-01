package particle

import (
	"io"
	"testing"

	"cleanroomorcontrol/internal/model"
)

// recordingInvalidator records the points it was asked to invalidate, so a test
// can assert that WriteSample drops the cached reading for the sampled point.
type recordingInvalidator struct {
	invalidated []model.PointID
	buckets     []model.PointID
}

func (r *recordingInvalidator) Invalidate(point model.PointID) {
	r.invalidated = append(r.invalidated, point)
}

func (r *recordingInvalidator) InvalidateBucket(point model.PointID) {
	r.buckets = append(r.buckets, point)
}

// memoryOpener returns in-memory write closers so tests never hold open file
// handles (which would block t.TempDir cleanup on Windows).
type memoryOpener struct{}

func (memoryOpener) Open(path string) (io.WriteCloser, error) {
	return &nopWriteCloser{}, nil
}

type nopWriteCloser struct{}

func (*nopWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (*nopWriteCloser) Close() error                { return nil }

func newTestMonitor(t *testing.T) (*Monitor, *recordingInvalidator) {
	t.Helper()
	table := NewTable()
	table.Register("P-01", "OR-1", 100)
	writer := NewRecordWriter(t.TempDir(), memoryOpener{})
	inv := &recordingInvalidator{}
	return NewMonitor(table, writer, inv), inv
}

// TestWriteSampleInvalidatesCache verifies the wiring fix for the bug: after a
// sample is written, the monitor must invalidate the cached reading for that
// point. Otherwise the alarm center reads a stale reading and misses the
// over-limit condition until the next cycle.
func TestWriteSampleInvalidatesCache(t *testing.T) {
	monitor, inv := newTestMonitor(t)

	sample := model.ParticleSample{
		Point:  "P-01",
		Room:   "OR-1",
		Count:  150,
		Volume: 1.0,
		At:     now(),
	}
	if err := monitor.WriteSample(sample); err != nil {
		t.Fatalf("WriteSample: %v", err)
	}

	if len(inv.invalidated) != 1 || inv.invalidated[0] != "P-01" {
		t.Fatalf("expected Invalidate(P-01), got %v", inv.invalidated)
	}

	// And the fresh reading is now visible through the table.
	reading := monitor.Reading("OR-1", "P-01")
	if !reading.HasData || reading.Count != 150 {
		t.Fatalf("expected fresh reading count 150, got %+v", reading)
	}
}

// TestWriteSampleNilCache guards the nil-safe path: a monitor built without an
// invalidator must not panic when writing a sample.
func TestWriteSampleNilCache(t *testing.T) {
	table := NewTable()
	table.Register("P-01", "OR-1", 100)
	monitor := NewMonitor(table, NewRecordWriter(t.TempDir(), memoryOpener{}), nil)

	sample := model.ParticleSample{
		Point:  "P-01",
		Room:   "OR-1",
		Count:  5,
		Volume: 1.0,
		At:     now(),
	}
	if err := monitor.WriteSample(sample); err != nil {
		t.Fatalf("WriteSample with nil cache: %v", err)
	}
}
