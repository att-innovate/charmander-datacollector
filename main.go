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
	"flag"
	"time"

	"github.com/golang/glog"
	"strconv"
	"os"
)

var config = Config{}
var argDbCreated = false

var PcpMetrics = []string{
	"cgroup.cpuacct.stat.user",
	"cgroup.cpuacct.stat.system",
	"cgroup.memory.usage",
	"network.interface.in.bytes",
	"network.interface.out.bytes",
	"network.interface.out.drops",
	"network.interface.in.drops",
}

func init() {
	//need to switch ip when deploying for prod vs local dev
	//var redisHost = flag.String("source_redis_host", "127.0.0.1:6379", "Redis IP Address:Port")
	flag.StringVar(
		&config.RedisHost,
		"source_redis_host",
		"172.31.2.11:31600",
		"Redis IP Address:Port",
	)

	flag.StringVar(
		&config.TorcHost,
		"source_torc_host",
		"wedge-fb-1:3000",
		"Redis IP Address:Port",
	)

	flag.StringVar(
		&config.Username,
		"influxdb_username",
		"root",
		"InfluxDB username",
	)

	flag.StringVar(
		&config.Password,
		"influxdb_password",
		"root",
		"InfluxDB password",
	)

	flag.StringVar(
		&config.DatabaseHost,
		"influxdb_host",
		"10.250.3.94:8086",
		"InfluxDB host:port",
	)

	flag.StringVar(
		&config.DatabaseName,
		"influxdb_name",
		"torc-datacollector",
		"Influxdb database name",
	)
	flag.StringVar(
		&config.Interval,
		"interval",
		"5",
		"Polling Interval: enter integer 1 - 5. Default is 5.",
	)
	flag.Parse()
}

func main() {
	var contextStore = NewContext()
	var hosts = getTorcNodes(config.TorcHost)
	var startTime = time.Now()
	for len(hosts) < 1 {
		glog.Error("Could not talk to torc to obtain host, retrying in 5 seconds.")
		time.Sleep(time.Second * 5)
		hosts = getTorcNodes(config.TorcHost)

		if time.Now().Sub(startTime) > 300*time.Second {
			glog.Fatal("Could not talk to torc to obtain host after 5 minutes, exiting.")
			os.Exit(1)
		}
	}

	contextStore.UpdateContext(hosts)
	//waits for all node to be up before starting work
	startTime = time.Now()
	for contextStore.Length() < len(hosts) {
		glog.Error("Could not reach pcp on host, retrying in 5 seconds.")
		time.Sleep(time.Second * 5)
		contextStore.UpdateContext(hosts)
		if time.Now().Sub(startTime) > 300*time.Second {
			glog.Fatal("Could not reach pcp on host to obtain context after 5 minutes, exiting.")
			os.Exit(1)
		}
	}
	
	GetInstanceMapping(contextStore)

	doWork(contextStore)
}

func doWork(contextStore *ContextList) {
	interval, err := strconv.Atoi(config.Interval)
	if err != nil {
		glog.Fatal("Unrecognized interval, Please use integers 1-5 Exiting.")
		os.Exit(1)
	}
	if interval <1 || interval >5 {
		interval = 5
		glog.Error("Interval outside of range, using 5 seconds.")
	}

	duration := time.Duration(interval)

	for host, _ := range contextStore.list {
		go func(host string, contextStore *ContextList) {
			for {
				var responseData = collectData(host, contextStore)
				if responseData.host != "" {
					processData(responseData)
				}
				time.Sleep(time.Second * duration)
			}
		}(host, contextStore)
	}

	keepAlive := make(chan int, 1)
	<-keepAlive
}
