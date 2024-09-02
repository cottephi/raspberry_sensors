package main

import (
	"log"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/conn/v3/physic"
)

type SensorData interface {
	Display() string
	Marshall() (string, error)
	UpdateFromEnv(env physic.Env)
}

type EnvSensorData struct {
	physic.Env
}

func (data *EnvSensorData) UpdateFromEnv(env physic.Env) {
	data.Env = env
}

type Sensor struct {
	Bus 			 string
	Address 	 uint16
	connection *bmxx80.Dev
	data 			 SensorData
}

func (sensor *Sensor) Start() {
	bus, err := i2creg.Open(sensor.Bus)

	if err != nil {
		log.Fatal(err)
	}

	sensor.connection, err = bmxx80.NewI2C(
		bus, sensor.Address, &bmxx80.DefaultOpts,
	)

	if err != nil {
		log.Fatal(err)
	}
}

func (sensor *Sensor) Read() {
	var env physic.Env
	if err := sensor.connection.Sense(&env); err != nil {
		log.Fatal(err)
	}
	sensor.data.UpdateFromEnv(env)
}

func (sensor *Sensor) Stop() {
	defer sensor.connection.Halt()
}