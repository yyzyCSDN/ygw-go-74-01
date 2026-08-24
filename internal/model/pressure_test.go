package model

import "testing"

func TestSummarizePressure(t *testing.T) {
	window := SummarizePressure("OR-1", []float64{20, 10, 30, 40})
	if window.Min != 10 || window.Max != 40 || window.Avg != 25 || window.Size != 4 {
		t.Fatalf("window = %+v", window)
	}
	empty := SummarizePressure("OR-1", nil)
	if empty.Size != 0 || empty.Min != 0 {
		t.Fatalf("empty window = %+v", empty)
	}
}

func TestPressureReadingNormalized(t *testing.T) {
	reading := PressureReading{Room: "OR-1", Pa: 22}
	normalized := reading.Normalized()
	if !normalized.Valid() {
		t.Fatal("normalized reading must be valid")
	}
	if normalized.Source != "sensor" {
		t.Fatalf("source = %q", normalized.Source)
	}
}

func TestParticleSummary(t *testing.T) {
	summary := SummarizeParticles("P-1", []ParticleReading{
		{Point: "P-1", Count: 50, Limit: 100, HasData: true},
		{Point: "P-1", Count: 150, Limit: 100, HasData: true},
		{Point: "P-1"},
	})
	if !summary.HasData || summary.Last != 150 || summary.Max != 150 || summary.Over != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}
