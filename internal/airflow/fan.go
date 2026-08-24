package airflow

import "cleanroomorcontrol/internal/model"

type FanDriver interface {
	Start(fan model.FanID) error
	Stop(fan model.FanID) error
}

func HealthyFans(units []model.FanUnit, role model.FanRole) []model.FanID {
	var fans []model.FanID
	for _, unit := range units {
		if unit.Role == role && unit.State == model.FanRunning {
			fans = append(fans, unit.ID)
		}
	}
	return fans
}

func UnitByID(units []model.FanUnit, id model.FanID) (model.FanUnit, bool) {
	for _, unit := range units {
		if unit.ID == id {
			return unit, true
		}
	}
	return model.FanUnit{}, false
}

func ActiveRoleFan(units []model.FanUnit, role model.FanRole) model.FanID {
	fans := HealthyFans(units, role)
	if len(fans) == 0 {
		return ""
	}
	return fans[0]
}
