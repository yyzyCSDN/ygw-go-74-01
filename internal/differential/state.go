package differential

import "cleanroomorcontrol/internal/model"

type PressureStateMachine struct {
	low          float64
	critical     float64
	droopNeed    int
	restoreNeed  int
	droopCount   int
	restoreCount int
	status       model.PressureStatus
	lastPa       float64
}

func NewPressureStateMachine(low, critical float64, droopNeed, restoreNeed int) *PressureStateMachine {
	return &PressureStateMachine{
		low:         low,
		critical:    critical,
		droopNeed:   droopNeed,
		restoreNeed: restoreNeed,
		status:      model.PressureStable,
	}
}

func (m *PressureStateMachine) Feed(pa float64) model.PressureStatus {
	m.lastPa = pa
	switch m.status {
	case model.PressureStable:
		if pa < m.low {
			m.droopCount++
			if m.droopCount >= m.droopNeed {
				m.droopCount = 0
				m.status = model.PressureDrooping
			}
		} else {
			m.droopCount = 0
		}
	case model.PressureDrooping:
		if pa < m.critical {
			m.status = model.PressureAlarm
			break
		}
		if pa >= m.low {
			m.restoreCount++
			if m.restoreCount >= m.restoreNeed {
				m.restoreCount = 0
				m.status = model.PressureStable
			}
		} else {
			m.restoreCount = 0
		}
	case model.PressureAlarm:
		if pa >= m.low {
			m.restoreCount++
			if m.restoreCount >= m.restoreNeed {
				m.restoreCount = 0
				m.status = model.PressureRestoring
			}
		} else {
			m.restoreCount = 0
		}
	case model.PressureRestoring:
		if pa >= m.low {
			m.status = model.PressureStable
		} else if pa < m.critical {
			m.status = model.PressureAlarm
		}
	}
	return m.status
}

func (m *PressureStateMachine) Status() model.PressureStatus {
	return m.status
}

func (m *PressureStateMachine) LastPa() float64 {
	return m.lastPa
}

func (m *PressureStateMachine) Reset() {
	m.status = model.PressureStable
	m.droopCount = 0
	m.restoreCount = 0
	m.lastPa = 0
}
