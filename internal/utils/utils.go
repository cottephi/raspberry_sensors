package utils

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"raspberry_sensors/internal/api"
	"raspberry_sensors/internal/logger"
	"raspberry_sensors/internal/sensors"
	"strings"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"periph.io/x/host/v3"
)

func StartLogger(logFile string) {
	err := logger.InitGlobalLogger(logFile, logger.DebugLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}
	defer logger.GlobalLogger.Close()

	if _, err := host.Init(); err != nil {
		logger.GlobalLogger.Fatal(err)
	}
}

func StartDB() *influxdb2.Client {

	INFLUXDB_SENSOR_TOKEN := os.Getenv("INFLUXDB_SENSOR_TOKEN")

	if INFLUXDB_SENSOR_TOKEN == "" {
		logger.GlobalLogger.Fatal("Environment variable 'INFLUXDB_SENSOR_TOKEN' not found or empty")
	}

	logger.GlobalLogger.Debug("Creating InfluxDB client...")
	influxClient := influxdb2.NewClient("http://localhost:8086", INFLUXDB_SENSOR_TOKEN)
	logger.GlobalLogger.Debug("...ok")
	return &influxClient
}

func StartSensors(influxClient *influxdb2.Client, dryRun bool) ([][2]chan bool, []*sensors.Sensor) {
	logger.GlobalLogger.Debug("Creating Sensors...")
	sensors_slice := []*sensors.Sensor{
		&sensors.NewSME280Sensor("/dev/i2c-1", 0x76, *influxClient, "raspberry", "seconds", dryRun).Sensor,
	}
	logger.GlobalLogger.Debug("...ok")

	var controlChannels [][2]chan bool

	for _, sensor := range sensors_slice {
		logger.GlobalLogger.Debugf("Starting sensor %s...", sensor.Name)
		sensorControlChannels := [2]chan bool{make(chan bool), make(chan bool)}

		sensor.Start()

		go sensor.Monitor(sensorControlChannels)
		controlChannels = append(controlChannels, sensorControlChannels)
		logger.GlobalLogger.Debug("...ok")
	}
	return controlChannels, sensors_slice
}

func StartServer(controlChannels [][2]chan bool) *api.Server {
	logger.GlobalLogger.Debug("Starting server...")
	server := api.NewServer(controlChannels)
	go server.Start(8080)
	logger.GlobalLogger.Debug("...ok")
	return server
}

func UseStartFlag(start bool) {
	if start {
		logger.GlobalLogger.Debug("Sending data acquisition start signal right away")
		err := QueryWithRetry("http://localhost:8080/sensors/start", 5*time.Second)
		if err != nil {
			logger.GlobalLogger.Fatalf("Failed to start sensors: %v", err)
		}
	}
}

func QueryWithRetry(url string, timeout time.Duration) error {
	// Deadline is the maximum time we allow for retries
	deadline := time.Now().Add(timeout)

	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			// Check if the error contains "connection refused"
			if strings.Contains(err.Error(), "connection refused") {
				lastErr = err
				time.Sleep(100 * time.Millisecond) // Short delay before retrying
				continue
			}
			return err // If it's a different error, return immediately
		}
		resp.Body.Close()
		return nil // Successful
	}
	// If we exit the loop, it means we exhausted the retries
	return lastErr
}

func WaitForExitSignal(server *api.Server) {
	// Create a channel to receive OS signals
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	defer close(server.QuitChan)

	select {
	case <-exitChan:
		logger.GlobalLogger.Info("Programm stopped. Shutting down...")
		resp, _ := http.Get("http://localhost:8080/sensors/stop")
		defer resp.Body.Close()
		logger.GlobalLogger.Info("Bye!")
	case <-server.QuitChan:
		// Nothing else to do, just acknowledge the channel
	}
}
