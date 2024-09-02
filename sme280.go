package main

import (
    "fmt"
		"time"
		"encoding/json"
)

type SME280Data struct {
	EnvSensorData
}

func (data SME280Data) Display() string {
	now := time.Now().Format(time.RFC3339)
	return fmt.Sprintf(
		"%s -- Temperature: %.2f°C, Pressure: %.2f hPa, Humidity: %.2f%%",
		now,
		float64(data.Temperature.Celsius()),
		float64(data.Pressure) / 1e11,  // nPa to hPa,
		float64(data.Humidity) / 1e5,  // tenth micro % to %
	)
}

func (data SME280Data) Marshall() (string, error) {
	now := time.Now().Unix()
	dictData := map[int64]map[string]float64{
		now: {
			"Temperature (K)": float64(data.Temperature) / 1e9,  // nK to K
			"Pressure (hPa)": float64(data.Pressure) / 1e11,
			"Humidity (%)": float64(data.Humidity) / 1e5,
		},
	}
	result, err := json.Marshal(dictData)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

type SME280Sensor struct {
	Sensor
}


func NewSME280Sensor(bus string, address uint16) *SME280Sensor {
	return &SME280Sensor{
		Sensor: Sensor{
			Bus:     bus,
			Address: address,
			data:    &SME280Data{},
			name:   "SME280",
		},
	}
}