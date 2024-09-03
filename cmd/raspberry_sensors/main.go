package main

import (
	"raspberry_sensors/internal/sensors"
	"raspberry_sensors/internal/api"
	"log"
	"periph.io/x/host/v3"
	"github.com/influxdata/influxdb-client-go/v2"
	"os"
	"os/signal"
	"flag"
	"net/http"
	"syscall"
)

func main() {

	start := flag.Bool("s", false, "If true, start data acquisition when program starts")
	flag.Parse()

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	INFLUXDB_SENSOR_TOKEN := os.Getenv("INFLUXDB_SENSOR_TOKEN")

	if INFLUXDB_SENSOR_TOKEN == "" {
		log.Fatal("Environment variable 'INFLUXDB_SENSOR_TOKEN' not found or empty")
	}

	influxClient := influxdb2.NewClient("http://localhost:8086", INFLUXDB_SENSOR_TOKEN)
	defer influxClient.Close()

	sensors_slice := []*sensors.Sensor {
		&sensors.NewSME280Sensor("/dev/i2c-1", 0x76, influxClient, "raspberry", "seconds").Sensor,
	}

	var controlChannels []chan bool

	for _, sensor := range sensors_slice {
		controlChannel := make(chan bool)
		defer close(controlChannel)

		sensor.Start()
		defer sensor.Stop()

		go sensor.Monitor(controlChannel)
		controlChannels = append(controlChannels, controlChannel)
	}

	server := api.NewServer(controlChannels)
	go server.Start(8080)

	if *start {
		resp, err := http.Get("http://localhost:8080/sensors/start")
		if err != nil {
			log.Fatalf("Failed to start sensors: %v", err)
		}
		defer resp.Body.Close()
	}

	waitForExitSignal(server)
}

func waitForExitSignal(server *api.Server) {
	// Create a channel to receive OS signals
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	defer close(server.QuitChan)

	select {
	case <-exitChan:
			log.Println("Shutting down...")
			resp, _ := http.Get("http://localhost:8080/sensors/stop")
			defer resp.Body.Close()
			log.Println("Bye!")
	case <-server.QuitChan:
		// Nothing else to do, just acknowledge the channel
	}
}