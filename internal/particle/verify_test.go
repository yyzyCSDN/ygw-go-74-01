package particle_test

import (
	"io"
	"testing"
	"time"

	"cleanroomorcontrol/internal/alarm"
	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type probeCloser struct{}

func (probeCloser) Write(p []byte) (int, error) { return len(p), nil }

func (probeCloser) Close() error { return nil }

type probeOpener struct{}

func (probeOpener) Open(path string) (io.WriteCloser, error) { return probeCloser{}, nil }

func TestAlarmCacheInvalidatedOnParticleWrite(t *testing.T) {
	at := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	table := particle.NewTable()
	table.Register("P-01", "OR-1", 100)
	table.Write("P-01", 50, 1.0, at)
	cache := alarm.NewCache(table)
	center := alarm.NewCenter(cache, nil)
	writer := particle.NewRecordWriter(t.TempDir(), probeOpener{})
	monitor := particle.NewMonitor(table, writer, cache)
	first := center.EvaluateParticlePoint("OR-1", "P-01")
	if first.Active {
		t.Fatal("low count must not alert")
	}
	if err := monitor.WriteSample(model.ParticleSample{Point: "P-01", Room: "OR-1", Count: 500, Volume: 1.0, At: at, Limit: 100}); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	second := center.EvaluateParticlePoint("OR-1", "P-01")
	if !second.Active {
		t.Fatal("stale cache: over-limit count not alerted")
	}
	if second.Count != 500 {
		t.Fatalf("count = %d, want 500", second.Count)
	}
}
