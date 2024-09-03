package sensors

import (
	"log"
	"time"
	"sync"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/conn/v3/physic"
)

type SensorData interface {
	Display()
	WriteToInfluxDB() error
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
	isRunning  bool
	mu         sync.Mutex // To handle concurrent access to isRunning
	name       string
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

func (sensor *Sensor) Read() error {
	var env physic.Env
	if err := sensor.connection.Sense(&env); err != nil {
		return err
	}
	sensor.data.UpdateFromEnv(env)
	return nil
}

func (sensor *Sensor) Stop() {
	defer sensor.connection.Halt() //nolint:errcheck
}

func (sensor *Sensor) Monitor(ctrlChan chan bool) {
	for { //nolint:gosimple
		select {
		case start := <-ctrlChan:
			if start {
				// Start monitoring if not already running
				sensor.mu.Lock()
				if !sensor.isRunning {
					log.Println("Starting sensor", sensor.name)
					sensor.isRunning = true
					sensor.mu.Unlock()

					go func() {
						for {
							sensor.mu.Lock()
							if !sensor.isRunning {
								sensor.mu.Unlock()
								return
							}
							sensor.mu.Unlock()

							if err := sensor.Read(); err != nil {
								log.Println(err)
							}	else {
								sensor.data.Display()
								if err := sensor.data.WriteToInfluxDB(); err != nil {
									log.Println(err)
								}
							}
							time.Sleep(time.Second)
						}
					}()
				} else {
					log.Println("Sensor", sensor.name, " already running")
					sensor.mu.Unlock()
				}
			} else {
				// Stop monitoring if running
				sensor.mu.Lock()
				if sensor.isRunning {
					sensor.isRunning = false
					log.Println("Stopped sensor", sensor.name)
				} else {
					log.Println("Sensor", sensor.name, "not running")
				}
				sensor.mu.Unlock()
			}
			ctrlChan <- true // Confirm that the sensor received its instruction
		}
	}
}