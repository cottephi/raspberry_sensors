package main

import (
	"flag"
	"raspberry_sensors/internal/config"
	"raspberry_sensors/internal/logger"
	"raspberry_sensors/internal/utils"

	"periph.io/x/host/v3"
)

func init() {
	c := config.Get()
	l := logger.Get()
	l.Info().Msg(c.Description)
	if _, err := host.Init(); err != nil {
		l.Fatal().Msgf("Failed to initialize periph.io: %v", err)
	}
}

func main() {

	start := flag.Bool("s", false, "If true, start data acquisition when program starts.")
	dryRun := flag.Bool("d", false, "Dry run. If True, will not write to DB.")
	flag.Parse()

	influxClient := utils.StartDB()
	if influxClient != nil {
		defer (*influxClient).Close()
	}

	controlChannels, sensors_slice := utils.StartSensors(influxClient, *dryRun)

	for i, sensor := range sensors_slice {
		defer close(controlChannels[i][0])
		defer close(controlChannels[i][1])
		defer sensor.Stop()
	}

	server := utils.StartServer(controlChannels)

	utils.UseStartFlag(*start)

	utils.WaitForExitSignal(server)
}
