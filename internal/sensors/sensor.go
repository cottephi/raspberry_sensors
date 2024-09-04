package sensors

import (
	"raspberry_sensors/internal/logger"
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
		logger.GlobalLogger.Fatal(err)
	}

	sensor.connection, err = bmxx80.NewI2C(
		bus, sensor.Address, &bmxx80.DefaultOpts,
	)

	if err != nil {
		logger.GlobalLogger.Fatal(err)
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

func (sensor *Sensor) Monitor(ctrlChans [2]chan bool) {
	for { //nolint:gosimple
		select {
		case start := <-ctrlChans[0]:
			if start {
				// Start monitoring if not already running
				sensor.mu.Lock()
				if !sensor.isRunning {
					logger.GlobalLogger.Infof("Starting sensor %s", sensor.name)
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
								logger.GlobalLogger.Error(err)
							}	else {
								sensor.data.Display()
								if err := sensor.data.WriteToInfluxDB(); err != nil {
									logger.GlobalLogger.Error(err)
								}
							}
							time.Sleep(time.Second)
						}
					}()
				} else {
					logger.GlobalLogger.Infof("Sensor %s already running", sensor.name)
					sensor.mu.Unlock()
				}
			} else {
				// Stop monitoring if running
				sensor.mu.Lock()
				if sensor.isRunning {
					sensor.isRunning = false
					logger.GlobalLogger.Infof("Stopped sensor %s", sensor.name)
				} else {
					logger.GlobalLogger.Infof("Sensor %s not running", sensor.name)
				}
				sensor.mu.Unlock()
			}
			ctrlChans[1] <- true // Confirm that the sensor received its instruction
		}
	}
}