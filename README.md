# Raspberry Sensors

This little project aims at operating various sensors on a Raspberry Pi, store their data in an InfluxDB database, and use Grafana to display them.

## Available sensors

* [BME280 Environment sensor](https://www.kubii.com/fr/modules-capteurs/2396-capteur-environnemental-kubii-3272496013285.html?mot_tcid=d79ebb19-daf2-4409-952b-452ce8c5958f) (temperature, pressure, humidity)

## Plugging the sensors

### Raspberry I2C ports

![Raspberry I2C ports](https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fwww.engineersgarage.com%2Fwp-content%2Fuploads%2F2020%2F08%2F25R-01.png&f=1&nofb=1&ipt=9f192155f75f00173b831680d90e328904b8e8df6dda137b1ca270f7eaabd658&ipo=images "Raspberry I2C ports")


### [Activate I2C connections](https://pi3g.com/enabling-and-checking-i2c-on-the-raspberry-pi-using-the-command-line-for-your-own-scripts/) on your Raspberry Pi

### BME280

Plug the wires on the Raspberry's I2C pins according to this table:

| Function Pin (BME280) | I2C Pin (Raspberry) |
|---|---|
| ADDR/MISO | GND |
| SCL/SCK | SCL |
| SDA/MOSI | SDA |
| GND | GND |
| VCC | 3.3V/5V |
| CS | not used |

(Found [here](https://www.waveshare.com/wiki/BME280_Environmental_Sensor))

## Installation

Assuming you plugged everything correctly

* [InfluxDB](https://pimylifeup.com/raspberry-pi-influxdb/)
* [Docker](https://docs.docker.com/engine/install/) (for Grafana, promtail and Loki)
* [Golang](https://pimylifeup.com/raspberry-pi-golang/)

Clone the reposotiry, `cd` in it and compile the code with

```bash
go build -o raspberry_sensors.exe ./cmd/raspberry_sensors/
```

## Configurations

Copy the following YAML configuration in a `config.yml` file at the root of the repository:


```yaml
api:
  # Overriden by env var SERVER_HOST, default is localhost
  host: "localhost"
  # Overriden by env var SERVER_PORT, default is 8080
  port: 8080

database:
  # Overriden by env var DB_HOST, default is http://localhost
  host: http://localhost
  # Overriden by env var DB_PORT, default is 8086
  port: 8086
  # Overriden by env var DB_TOKEN, no default, mandatory
  token: <your influx DB token>

logger:
  # zerolog levels are 'trace', 'debug', 'info', 'warn', 'error', 'fatal' and 'panic'
  level: "info"
```

To test the sensors code, you do not need to give InfluxDB's token at first. Instead, run

```bash
./raspberry_sensors.exe -d -s
```

The `-d` option indicates a dry run, which will only output the result in the console, and note write anything to the database. The `-s` option indicates that the data acquisition should start right away. If everything is normal, you should see:

```bash
No logfile path provided, only logging in the console
2024-09-08T11:59:01+02:00 INF cmd/raspberry_sensors/main.go:15 > Configuration:
 - Server Host: http://localhost
 - Server Port: 8080
 - Logger Level: INFO
 - Log file path: 
 - Server URL: http://localhost:8080
No database token given, not writing data to database
2024-09-08T11:59:01+02:00 WRN internal/utils/utils.go:25 > No InfluxDB token provided. Not starting Database.
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:42 > Creating Sensors...
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:46 > ...ok
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:52 > Starting sensor BME280...
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:59 > ...ok
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:66 > Starting server...
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:69 > ...ok
2024-09-08T11:59:01+02:00 INF internal/api/server.go:61 > Listening/Serving on localhost:8080
2024-09-08T11:59:01+02:00 INF internal/utils/utils.go:77 > Sending data acquisition start signal right away
2024-09-08T11:59:01+02:00 INF internal/sensors/sensor.go:84 > Starting data acquisition of sensor BME280
2024-09-08T11:59:01+02:00 INF internal/sensors/bme280.go:17 > Temperature: 25.87°C, Pressure: 995.90 hPa, Humidity: 52.23%
2024-09-08T11:59:02+02:00 INF internal/sensors/bme280.go:17 > Temperature: 25.88°C, Pressure: 995.88 hPa, Humidity: 52.18%
2024-09-08T11:59:03+02:00 INF internal/sensors/bme280.go:17 > Temperature: 25.88°C, Pressure: 995.91 hPa, Humidity: 52.16%
```

## Database configuration

Go to http://localhost:8086 and login to InfluxDB. Then do the following

* Create an organisation named "raspberry"
* Create a bucket named "seconds" with a retention policy of 2 days
* Create a bucket named "daily". Choose the retention policy (I did not use one)
* Import the task found in *internal/database/hourly_aggregation.json* in InfluxDB Tasks.
* Create a token for your sensor, or use the admin one (not recommended). It only needs read/write access to the 'seconds' bucket.

You can provide your InfluxDB token to the configuration and restart the program without the `-d` option.

## Logging configuration

The program can also write logs to a file that Grafana/Loki will process. To do that, run the script `setup_log_dirs.sh`. That will create the **loki** directory in your home folder, and create a local .env file that the program will detect, and from which the path where the logs should be written will be read. It is **$HOME/logs/raspberry_sensors**. It will be created when you first run the program.

Run the main program, then in another terminal in the repository directory run

```bash
docker compose up -d
```

This will start Grafana, Promtail and Loki. Check their logs to see if everything looks good (`docker logs grafana/loki/promtail`) and go to http://localhost:3000 and log into Grafana with *admin:admin*. Then go to Dashboard -> New -> Import and use the file *internal/grafana/Logs-dashboard.json*. You might have to setup you data source to **Loki** first.

The log files are rotating files of at most 1GB each, keeping at most the last 5 log files for at most 7 days.

## Service configuration

To make this program start with your Raspberry Pi, create a Systemd service :

```bash
sudo nano /etc/systemd/system/raspberry_sensors.service
```

with the following content

```
[Unit]
Description=Raspberry Sensors Go Program
After=influxdb.service

[Service]
WorkingDirectory=<your home>/projects/raspberry_sensors
ExecStart=<your home>/projects/raspberry_sensors/raspberry_sensors.exe -s
Restart=always
User=<your user name>

[Install]
WantedBy=multi-user.target
```

and a second one

```bash
sudo nano /etc/systemd/system/docker-compose-monitoring.service
```

with the following content

```
[Unit]
Description=Docker Compose Stack for Grafana, Promtail, Loki
Requires=docker.service
After=docker.service

[Service]
WorkingDirectory=<your home>/projects/raspberry_sensors
ExecStart=/usr/bin/docker compose up
ExecStop=/usr/bin/docker compose down
Restart=always
User=<your user name>

[Install]
WantedBy=multi-user.target
```

Run then

```bash
sudo systemctl daemon-reload
sudo systemctl enable docker-compose-monitoring.service
sudo systemctl enable raspberry_sensors.service
sudo systemctl start docker-compose-monitoring.service
sudo systemctl start raspberry_sensors.service
```


## Useful commands

Empty a bucket in the database

```bash
influx delete --bucket <bucket> --predicate '_measurement="environment_bme280"' --start '1970-01-01T00:00:00Z' --stop $(date +"%Y-%m-%dT%H:%M:%SZ") --org raspberry --token <admin token>
```

Stop data acquisition without killing the program

```bash
wget -qO- localhost:8080/sensors/stop
```

Start data acquisition (if you stopped it with the previous command or did not start the program with the `-s` flag)

```bash
wget -qO- localhost:8080/sensors/start
```

Kill the program

```bash
wget -qO- localhost:8080/sensors/kill
```

Note that the API routes can be available from any computer on your local network if you know your Raspberry Pi's IP address.