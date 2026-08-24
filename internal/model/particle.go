package model

import "time"

type ParticleSample struct {
	Point  PointID
	Room   RoomID
	Count  int
	Volume float64
	At     time.Time
	Limit  int
}

func (s ParticleSample) Valid() bool {
	return s.Point != "" && s.Room != "" && s.Volume > 0
}

func (s ParticleSample) Normalized() ParticleSample {
	if s.At.IsZero() {
		s.At = time.Now()
	}
	return s
}

type ParticleReading struct {
	Point   PointID
	Room    RoomID
	Count   int
	Volume  float64
	At      time.Time
	Limit   int
	HasData bool
}

func (r ParticleReading) OverLimit() bool {
	return r.HasData && r.Count > r.Limit
}

type ParticleSummary struct {
	Point   PointID
	Samples int
	Max     int
	Last    int
	Over    int
	HasData bool
}

func SummarizeParticles(point PointID, samples []ParticleReading) ParticleSummary {
	summary := ParticleSummary{Point: point, Samples: len(samples)}
	for _, sample := range samples {
		if !sample.HasData {
			continue
		}
		summary.HasData = true
		summary.Last = sample.Count
		if sample.Count > summary.Max {
			summary.Max = sample.Count
		}
		if sample.OverLimit() {
			summary.Over++
		}
	}
	return summary
}
