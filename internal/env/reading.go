package env

import "cleanroomorcontrol/internal/model"

type ReadingAggregate struct {
	Room        model.RoomID
	Count       int
	TempMax     float64
	TempMin     float64
	HumidityMax float64
	HumidityMin float64
}

func Aggregate(room model.RoomID, readings []model.EnvReading) ReadingAggregate {
	aggregate := ReadingAggregate{Room: room, Count: len(readings)}
	if len(readings) == 0 {
		return aggregate
	}
	aggregate.TempMin = readings[0].TempC
	aggregate.TempMax = readings[0].TempC
	aggregate.HumidityMin = readings[0].Humidity
	aggregate.HumidityMax = readings[0].Humidity
	for _, reading := range readings[1:] {
		if reading.TempC < aggregate.TempMin {
			aggregate.TempMin = reading.TempC
		}
		if reading.TempC > aggregate.TempMax {
			aggregate.TempMax = reading.TempC
		}
		if reading.Humidity < aggregate.HumidityMin {
			aggregate.HumidityMin = reading.Humidity
		}
		if reading.Humidity > aggregate.HumidityMax {
			aggregate.HumidityMax = reading.Humidity
		}
	}
	return aggregate
}
