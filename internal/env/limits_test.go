package env

import (
	"testing"

	"cleanroomorcontrol/internal/model"
)

func TestLimitsViolations(t *testing.T) {
	limits := model.EnvLimits{TempMin: 18, TempMax: 26, HumidityMin: 35, HumidityMax: 65}
	reading := model.EnvReading{Room: "OR-1", TempC: 28, Humidity: 30}
	violations := limits.Violations(reading)
	if len(violations) != 2 {
		t.Fatalf("violations = %v", violations)
	}
	clean := model.EnvReading{Room: "OR-1", TempC: 22, Humidity: 50}
	if len(limits.Violations(clean)) != 0 {
		t.Fatal("clean reading must have no violations")
	}
}

func TestLimitsStore(t *testing.T) {
	store := NewLimitsStore()
	if _, ok := store.Get("OR-1"); ok {
		t.Fatal("unregistered room must not resolve")
	}
	store.Set("OR-1", model.EnvLimits{TempMin: 18, TempMax: 26})
	limits, ok := store.Get("OR-1")
	if !ok || limits.TempMax != 26 {
		t.Fatalf("limits = %+v ok=%v", limits, ok)
	}
}

func TestEnvReadingNormalized(t *testing.T) {
	reading := model.EnvReading{Room: "OR-1", Humidity: 120}
	normalized := reading.Normalized()
	if normalized.Humidity != 100 {
		t.Fatalf("humidity = %v", normalized.Humidity)
	}
}
