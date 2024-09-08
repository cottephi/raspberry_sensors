package utils

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"raspberry_sensors/internal/api"
	"raspberry_sensors/internal/config"
	"raspberry_sensors/internal/logger"
	"raspberry_sensors/internal/sensors"
	"strings"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func StartDB() *influxdb2.Client {

	c := config.Get()
	l := logger.Get()

	if c.Database.Token == "" {
		l.Warn().Msg("No InfluxDB token provided. Not starting Database.")
		return nil
	}

	l.Info().Msg("Creating InfluxDB client...")
	influxClient := influxdb2.NewClient(
		fmt.Sprintf(
			"%s:%s", c.Database.Host, c.Database.Port,
		),
		c.Database.Token,
	)
	l.Info().Msg("...ok")
	return &influxClient
}

func StartSensors(influxClient *influxdb2.Client, dryRun bool) ([][2]chan bool, []*sensors.Sensor) {
	l := logger.Get()
	l.Info().Msg("Creating Sensors...")
	sensors_slice := []*sensors.Sensor{
		&sensors.NewBME280Sensor("/dev/i2c-1", 0x76, influxClient, "raspberry", "seconds", dryRun).Sensor,
	}
	l.Info().Msg("...ok")

	var controlChannels [][2]chan bool

	for _, sensor := range sensors_slice {
		l := logger.Get()
		l.Info().Msgf("Starting sensor %s...", sensor.Name)
		sensorControlChannels := [2]chan bool{make(chan bool), make(chan bool)}

		sensor.Start()

		go sensor.Monitor(sensorControlChannels)
		controlChannels = append(controlChannels, sensorControlChannels)
		l.Info().Msg("...ok")
	}
	return controlChannels, sensors_slice
}

func StartServer(controlChannels [][2]chan bool) *api.Server {
	l := logger.Get()
	l.Info().Msg("Starting server...")
	server := api.NewServer(controlChannels)
	go server.Start()
	l.Info().Msg("...ok")
	return server
}

func UseStartFlag(start bool) {
	if start {
		l := logger.Get()
		c := config.Get()
		l.Info().Msg("Sending data acquisition start signal right away")
		err := QueryWithRetry(
			fmt.Sprintf(
				"%s:%s/sensors/start", c.Api.Host, c.Api.Port,
			),
			5*time.Second,
		)
		if err != nil {
			l.Fatal().Msgf("Failed to start sensors: %v", err)
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
	l := logger.Get()
	c := config.Get()
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	defer close(server.QuitChan)

	select {
	case <-exitChan:
		l.Info().Msg("Programm stopped. Shutting down...")
		resp, _ := http.Get(
			fmt.Sprintf(
				"%s:%s/sensors/stop", c.Api.Host, c.Api.Port,
			),
		)
		if resp != nil {
			defer resp.Body.Close()
		}
		l.Info().Msg("Bye!")
	case <-server.QuitChan:
		// Nothing else to do, just acknowledge the channel
	}
}
