package main

import (
	"raspberry_sensors/internal/sensors"
	"raspberry_sensors/internal/api"
	"raspberry_sensors/internal/utils"
	"raspberry_sensors/internal/logger"
	"periph.io/x/host/v3"
	"github.com/influxdata/influxdb-client-go/v2"
	"os"
	"flag"
	"fmt"
	"time"
)

func main() {

	start := flag.Bool("s", false, "If true, start data acquisition when program starts")
	logFile := flag.String("l", "", "File to write logs to. If not specified, will not write logs to file.")
	flag.Parse()

	err := logger.InitGlobalLogger(*logFile, logger.DebugLevel)
	if err != nil {
			fmt.Printf("Failed to initialize logger: %v", err)
			os.Exit(1)
	}

	if _, err := host.Init(); err != nil {
		logger.GlobalLogger.Fatal(err)
	}

	INFLUXDB_SENSOR_TOKEN := os.Getenv("INFLUXDB_SENSOR_TOKEN")

	if INFLUXDB_SENSOR_TOKEN == "" {
		logger.GlobalLogger.Fatal("Environment variable 'INFLUXDB_SENSOR_TOKEN' not found or empty")
	}

	influxClient := influxdb2.NewClient("http://localhost:8086", INFLUXDB_SENSOR_TOKEN)
	defer influxClient.Close()

	sensors_slice := []*sensors.Sensor {
		&sensors.NewSME280Sensor("/dev/i2c-1", 0x76, influxClient, "raspberry", "seconds").Sensor,
	}

	var controlChannels [][2]chan bool

	for _, sensor := range sensors_slice {
		sensorControlChannels := [2]chan bool{make(chan bool), make(chan bool)}
		defer close(sensorControlChannels[0])
		defer close(sensorControlChannels[1])

		sensor.Start()
		defer sensor.Stop()

		go sensor.Monitor(sensorControlChannels)
		controlChannels = append(controlChannels, sensorControlChannels)
	}

	server := api.NewServer(controlChannels)
	go server.Start(8080)

	if *start {
		err := utils.QueryWithRetry("http://localhost:8080/sensors/start", 5 * time.Second)
    if err != nil {
			logger.GlobalLogger.Fatalf("Failed to start sensors: %v", err)
    }
	}
	utils.WaitForExitSignal(server)
}