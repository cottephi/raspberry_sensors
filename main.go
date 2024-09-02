package main

import (
	"log"
	"periph.io/x/host/v3"
)

func main() {

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	sme280 := NewSME280Sensor("/dev/i2c-1", 0x76)

	sme280.Start()
	defer sme280.Stop()

	sme280.Read()

	// Display the sensor data
	log.Println(sme280.data.Display())

	// Marshall the sensor data to JSON
	jsonData, err := sme280.data.Marshall()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("JSON Data:", jsonData)

}
