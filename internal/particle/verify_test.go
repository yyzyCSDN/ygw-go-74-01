package particle_test

import (
	"io"
	"testing"
	"time"

	"cleanroomorcontrol/internal/alarm"
	"cleanroomorcontrol/internal/model"
	"cleanroomorcontrol/internal/particle"
)

type countingOpener struct {
	opened int
	closed int
}

type countingCloser struct {
	owner *countingOpener
}

func (c *countingCloser) Write(p []byte) (int, error) { return len(p), nil }

func (c *countingCloser) Close() error {
	c.owner.closed++
	return nil
}

func (c *countingOpener) Open(path string) (io.WriteCloser, error) {
	c.opened++
	return &countingCloser{owner: c}, nil
}

func TestParticleFileHandleClosed(t *testing.T) {
	opener := &countingOpener{}
	table := particle.NewTable()
	table.Register("P-01", "OR-1", 100)
	cache := alarm.NewCache(table)
	writer := particle.NewRecordWriter(t.TempDir(), opener)
	monitor := particle.NewMonitor(table, writer, cache)
	for i := 0; i < 3; i++ {
		if err := monitor.WriteSample(model.ParticleSample{Point: "P-01", Room: "OR-1", Count: 10 + i, Volume: 1.0, At: time.Now(), Limit: 100}); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}
	if err := monitor.RotateRecord("OR-1"); err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	if opener.opened != 1 || opener.closed != 1 {
		t.Fatalf("handles opened=%d closed=%d, want 1 and 1", opener.opened, opener.closed)
	}
}
