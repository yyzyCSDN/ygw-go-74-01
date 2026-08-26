package model

import "time"

type EnvReading struct {
	Room     RoomID
	TempC    float64
	Humidity float64
	At       time.Time
}

func (r EnvReading) Valid() bool {
	return r.Room != "" && !r.At.IsZero()
}

func (r EnvReading) Normalized() EnvReading {
	if r.At.IsZero() {
		r.At = time.Now()
	}
	if r.Humidity < 0 {
		r.Humidity = 0
	}
	if r.Humidity > 100 {
		r.Humidity = 100
	}
	return r
}

type EnvLimits struct {
	TempMin     float64
	TempMax     float64
	HumidityMin float64
	HumidityMax float64
}

func (l EnvLimits) TempViolations(tempC float64) []string {
	var violations []string
	if tempC < l.TempMin {
		violations = append(violations, "temperature-below-min")
	}
	if tempC > l.TempMax {
		violations = append(violations, "temperature-above-max")
	}
	return violations
}

func (l EnvLimits) HumidityViolations(humidity float64) []string {
	var violations []string
	if humidity < l.HumidityMin {
		violations = append(violations, "humidity-below-min")
	}
	if humidity > l.HumidityMax {
		violations = append(violations, "humidity-above-max")
	}
	return violations
}

func (l EnvLimits) Violations(reading EnvReading) []string {
	violations := l.TempViolations(reading.TempC)
	violations = append(violations, l.HumidityViolations(reading.Humidity)...)
	return violations
}
