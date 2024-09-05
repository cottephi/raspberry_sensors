package sensors

import (
	"context"
	"raspberry_sensors/internal/logger"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
)

type BME280Data struct {
	EnvSensorData
}

func (data BME280Data) Display() {
	l := logger.Get()
	l.Info().Msgf(
		"Temperature: %.2f°C, Pressure: %.2f hPa, Humidity: %.2f%%",
		float64(data.Temperature.Celsius()),
		float64(data.Pressure)/1e11, // nPa to hPa,
		float64(data.Humidity)/1e5,  // tenth micro % to %
	)
}

func (data BME280Data) WriteToInfluxDB() error {
	l := logger.Get()
	if data.dryRun {
		l.Debug().Msg("Dry run, not writing to DB")
		return nil
	}
	l.Debug().Msgf("Writing to DB: Org=%s, Bucket=%s, measurement=%s, tags=%s", data.Org, data.Bucket, data.name, data.tags)
	now := time.Now()

	writeAPI := data.InfluxClient.WriteAPIBlocking(data.Org, data.Bucket)

	p := influxdb2.NewPoint(
		data.name, // measurement name
		data.tags, // tags
		map[string]interface{}{
			"temperature_celsius": float64(data.Temperature.Celsius()),
			"pressure_hpa":        float64(data.Pressure) / 1e11,
			"humidity_percent":    float64(data.Humidity) / 1e5,
		}, // fields
		now, // timestamp
	)

	// Write the point to the database
	return writeAPI.WritePoint(context.Background(), p)
}

type BME280Sensor struct {
	Sensor
}

func NewBME280Sensor(bus string, address uint16, influxClient influxdb2.Client, org, bucket string, dryRun bool) *BME280Sensor {
	return &BME280Sensor{
		Sensor: Sensor{
			Bus:     bus,
			Address: address,
			data: &BME280Data{
				EnvSensorData: EnvSensorData{
					InfluxClient: influxClient,
					Org:          org,
					Bucket:       bucket,
					tags:         map[string]string{"sensor": "BME280", "location": "office"},
					name:         "environment_bme280",
					dryRun:       dryRun,
				},
			},
			Name: "BME280",
		},
	}
}
