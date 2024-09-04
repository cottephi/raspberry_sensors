package sensors

import (
	"raspberry_sensors/internal/logger"
	"time"
	"context"
	"github.com/influxdata/influxdb-client-go/v2"
)

type SME280Data struct {
	EnvSensorData
}

func (data SME280Data) Display() {
	logger.GlobalLogger.Infof(
		"Temperature: %.2f°C, Pressure: %.2f hPa, Humidity: %.2f%%\n",
		float64(data.Temperature.Celsius()),
		float64(data.Pressure) / 1e11,  // nPa to hPa,
		float64(data.Humidity) / 1e5,  // tenth micro % to %
	)
}

func (data SME280Data) WriteToInfluxDB() error {
	if data.dryRun {
		logger.GlobalLogger.Debug("Dry run, not writing to DB")
		return nil	
	}
	logger.GlobalLogger.Debugf("Writing to DB: Org=%s, Bucket=%s, measurement=%s, tags=%s", data.Org, data.Bucket, data.name, data.tags)
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

type SME280Sensor struct {
	Sensor
}


func NewSME280Sensor(bus string, address uint16, influxClient influxdb2.Client, org, bucket string, dryRun bool) *SME280Sensor {
	return &SME280Sensor{
		Sensor: Sensor{
			Bus:     bus,
			Address: address,
			data:    &SME280Data{
				EnvSensorData: EnvSensorData{
					InfluxClient: influxClient,
					Org:          org,
					Bucket:       bucket,
					tags:         map[string]string{"sensor": "SME280", "location": "office"},
					name:         "environment_sme280",
					dryRun: 		  dryRun,
				},
			},
			Name:   "SME280",
		},
	}
}