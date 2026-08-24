package alarm

import (
	"testing"

	"cleanroomorcontrol/internal/model"
)

func TestSeverityPressure(t *testing.T) {
	if severityPressure(model.PressureAlarm) != "critical" {
		t.Fatal("alarm pressure must be critical")
	}
	if severityPressure(model.PressureDrooping) != "warning" {
		t.Fatal("drooping pressure must be warning")
	}
	if severityPressure(model.PressureStable) != "normal" {
		t.Fatal("stable pressure must be normal")
	}
}

func TestSeverityParticle(t *testing.T) {
	if severityParticle(ParticleAlert{Active: true}) != "critical" {
		t.Fatal("active particle alert must be critical")
	}
	if severityParticle(ParticleAlert{Active: false}) != "normal" {
		t.Fatal("inactive particle alert must be normal")
	}
}

func TestSeverityEnv(t *testing.T) {
	if severityEnv([]string{"temperature-above-max"}) != "warning" {
		t.Fatal("env violation must be warning")
	}
	if severityEnv(nil) != "normal" {
		t.Fatal("no violation must be normal")
	}
}

func TestFormatPa(t *testing.T) {
	if formatPa(-1) != "negative" || formatPa(0) != "zero" || formatPa(3) != "positive" {
		t.Fatal("formatPa classification failed")
	}
}
