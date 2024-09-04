package main

import (
	"flag"
	"raspberry_sensors/internal/utils"
)

func main() {

	start := flag.Bool("s", false, "If true, start data acquisition when program starts")
	logFile := flag.String("l", "", "File to write logs to. If not specified, will not write logs to file.")
	dryRun := flag.Bool("d", false, "Dry run. If True, will not write to DB")
	flag.Parse()

	utils.StartLogger(*logFile)

	influxClient := utils.StartDB()
	defer (*influxClient).Close()

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
