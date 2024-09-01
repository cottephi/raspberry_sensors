package main

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

func main() {
	// Initialize periph.io host
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a handle to the I2C bus
	bus, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new connection to the SME280 sensor
	dev, err := bmxx80.NewI2C(bus, 0x76, &bmxx80.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Halt()

	// Read and print data in a loop
	for {
		// Read sensor data
		var env physic.Env
		if err := dev.Sense(&env); err != nil {
			log.Fatal(err)
		}

		// Print temperature, pressure, and humidity
		fmt.Printf("Temperature: %s\n", env.Temperature)
		fmt.Printf("Pressure:    %s\n", env.Pressure)
		fmt.Printf("Humidity:    %s\n", env.Humidity)

		time.Sleep(2 * time.Second)
	}
}

