package main

import (
	"raspberry_sensors/internal/sensors"
	"raspberry_sensors/internal/api"
	"log"
	"periph.io/x/host/v3"
	"github.com/influxdata/influxdb-client-go/v2"
	"os"
)

func main() {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	ctrlChan := make(chan bool)
	defer close(ctrlChan)

	INFLUXDB_SENSOR_TOKEN := os.Getenv("INFLUXDB_SENSOR_TOKEN")

	if INFLUXDB_SENSOR_TOKEN == "" {
		log.Fatal("Environment variable 'INFLUXDB_SENSOR_TOKEN' not found or empty")
	}

	influxClient := influxdb2.NewClient("http://localhost:8086", INFLUXDB_SENSOR_TOKEN)
	defer influxClient.Close()

	sensors_slice:= []*sensors.Sensor {
		&sensors.NewSME280Sensor("/dev/i2c-1", 0x76, influxClient, "raspberry", "seconds").Sensor,
	}

	for _, sensor := range sensors_slice {
		sensor.Start()
		defer sensor.Stop()
		go sensor.Monitor(ctrlChan)
	}

	server := api.NewServer(ctrlChan)
	server.Start(8080)
	
}
