package main

import (
	"log"
	"periph.io/x/host/v3"
)

func main() {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	ctrlChan := make(chan bool)
	defer close(ctrlChan)

	sensors:= []*Sensor {
		&NewSME280Sensor("/dev/i2c-1", 0x76).Sensor,
	}

	for _, sensor := range sensors {
		sensor.Start()
		defer sensor.Stop()
		go sensor.Monitor(ctrlChan)
	}

	server := NewServer(ctrlChan)
	server.Start(8080)
	
}
