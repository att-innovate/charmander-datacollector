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
	"page_faults",
	"cpu_usage_system",
	"memory_working_set",
	"cpu_usage_user",
	}

var Stats = []string{
	"time",
	"memory_usage",
	"page_faults",
	"hostname",
	"container_name",
	"cpu_usage_user",
	"memory_working_set",
	"cpu_usage_system",
	//"network_in_bytes",
	//"network_out_bytes",
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

	if dataType == "machine" {
		series := &client.Series{
			Name:    "machine",
			Columns: Machine,
			Points:  data,

		}

		if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
			fmt.Println("failed to write stats to influxDb - %s", err)
		}
	}

	if dataType == "stats" {
		series := &client.Series{
			Name:    "stats",
			Columns: Stats,
			Points:  data,
		}

		if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
			fmt.Println("failed to write stats to influxDb - %s", err)
		}
	}

}
