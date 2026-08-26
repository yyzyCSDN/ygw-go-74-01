package model

import "time"

type PressureReading struct {
	Room   RoomID
	Pa     float64
	At     time.Time
	Source string
}

func (r PressureReading) Valid() bool {
	return r.Room != "" && !r.At.IsZero()
}

func (r PressureReading) Normalized() PressureReading {
	if r.At.IsZero() {
		r.At = time.Now()
	}
	if r.Source == "" {
		r.Source = "sensor"
	}
	return r
}

type PressureWindow struct {
	Room RoomID
	Min  float64
	Max  float64
	Avg  float64
	Size int
}

func SummarizePressure(room RoomID, values []float64) PressureWindow {
	window := PressureWindow{Room: room, Size: len(values)}
	if len(values) == 0 {
		return window
	}
	window.Min = values[0]
	window.Max = values[0]
	total := 0.0
	for _, value := range values {
		if value < window.Min {
			window.Min = value
		}
		if value > window.Max {
			window.Max = value
		}
		total += value
	}
	window.Avg = total / float64(len(values))
	return window
}
