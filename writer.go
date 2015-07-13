package main

import (
	"flag"
	"fmt"
	"github.com/influxdb/influxdb/client"
)

var Machine = []string{
	"time",
	"hostname",
	"memory_usage",
	"cpu_usage_system",
	"cpu_usage_user",
	}

var Stats = []string{
	"time",
	"memory_usage",
	"hostname",
	"container_name",
	"cpu_usage_user",
	"cpu_usage_system",
	}

var Network = []string{
	"time",
	"hostname",
	"network_in_bytes",
	"network_out_bytes",
	"interface_name",
	"network_in_drops",
	"network_out_drops",
}

var (
	argDbUsername = flag.String("sink_influxdb_username", "root", "InfluxDB username")
	argDbPassword = flag.String("sink_influxdb_password", "root", "InfluxDB password")
	argDbHost     = flag.String("sink_influxdb_host", "172.31.2.11:31410", "InfluxDB host:port")
	argDbName     = flag.String("sink_influxdb_name", "test2", "Influxdb database name")
	argDbCreated = false
)

func Write(data [][]interface{}, dataType string) {

	c, err := client.NewClient(&client.ClientConfig{
		Host:     *argDbHost,
		Username: *argDbUsername,
		Password: *argDbPassword,
		Database: *argDbName,
	})

	if (argDbCreated == false) {
		argDbCreated = true
		if err := c.CreateDatabase(*argDbName); err != nil {
			fmt.Println("Database creating error - %s", err)
		}
	}

	c.DisableCompression()

	if err != nil {
		panic(err)
	}

	var column []string

	switch dataType {
	case "machine":
		column = Machine
	case "stats":
		column = Stats
	case "network":
		column = Network
	default:
		fmt.Println("Error: unrecognized database")
		return
	}

	series := &client.Series{
		Name:    dataType,
		Columns: column,
		Points:  data,
	}

	if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
		fmt.Println("failed to write ",dataType," to influxDb - %s", err)
	}

}