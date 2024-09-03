package main

import (
	"raspberry_sensors/internal/sensors"
	"raspberry_sensors/internal/api"
	"log"
	"periph.io/x/host/v3"
)

func main() {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	ctrlChan := make(chan bool)
	defer close(ctrlChan)

	sensors_slice:= []*sensors.Sensor {
		&sensors.NewSME280Sensor("/dev/i2c-1", 0x76).Sensor,
	}

	for _, sensor := range sensors_slice {
		sensor.Start()
		defer sensor.Stop()
		go sensor.Monitor(ctrlChan)
	}

	server := api.NewServer(ctrlChan)
	server.Start(8080)
	
}
