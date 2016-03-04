// The MIT License (MIT)
//
// Copyright (c) 2014 AT&T
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"
	"github.com/influxdb/influxdb/client"
	"strings"
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

func Write(data [][]interface{}, dataType string) bool {

	c, err := client.NewClient(&client.ClientConfig{
		Host:     config.DatabaseHost,
		Username: config.Username,
		Password: config.Password,
		Database: config.DatabaseName,
	})

	if argDbCreated == false {
		argDbCreated = true
		if err := c.CreateDatabase(config.DatabaseName); err != nil {
			fmt.Println("Error creating database:", err)
			fmt.Println("Using existing database.")
		} else {
			fmt.Println("Creating Database")
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
		return false
	}

	series := &client.Series{
		Name:    dataType,
		Columns: column,
		Points:  data,
	}

	if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
		fmt.Println("Failed to write", dataType, "to influxDb.", err)
		fmt.Println("Data:", series)
		if strings.Contains(err.Error(), "400") {
			argDbCreated = false
		}
		return false
	}

	return true
}
