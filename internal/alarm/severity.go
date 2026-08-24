package alarm

import "cleanroomorcontrol/internal/model"

func severityPressure(status model.PressureStatus) string {
	switch status {
	case model.PressureAlarm:
		return "critical"
	case model.PressureDrooping:
		return "warning"
	case model.PressureRestoring:
		return "info"
	default:
		return "normal"
	}
}

func severityParticle(alert ParticleAlert) string {
	if alert.Active {
		return "critical"
	}
	return "normal"
}

func severityEnv(violations []string) string {
	if len(violations) > 0 {
		return "warning"
	}
	return "normal"
}

func formatPa(pa float64) string {
	if pa < 0 {
		return "negative"
	}
	if pa == 0 {
		return "zero"
	}
	return "positive"
}
