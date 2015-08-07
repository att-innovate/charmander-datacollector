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
		"172.31.2.11:31410",
		"InfluxDB host:port",
	)

	flag.StringVar(
		&config.DatabaseName,
		"influxdb_name",
		"charmander-dc",
		"Influxdb database name",
	)
	flag.IntVar(
		&config.Interval,
		"interval",
		5,
		"Polling Interval: enter integer 1 - 5. Default is 5.",
	)
	flag.Parse()
}

func main() {
	var contextStore = NewContext()
	var hosts = GetCadvisorHosts()
	var startTime = time.Now()
	for len(hosts) < 1 {
		glog.Error("Could not talk to redis to obtain host, retrying in 5 seconds.")
		time.Sleep(time.Second * 5)
		hosts = GetCadvisorHosts()

		if time.Now().Sub(startTime) > 300*time.Second {
			glog.Fatal("Could not talk to redis to obtain host after 5 minutes, exiting.")
		}
	}

	contextStore.UpdateContext(hosts)

	GetInstanceMapping(contextStore)

	doWork(contextStore)
}

func doWork(contextStore *ContextList) {
	if (config.Interval < 1){
		config.Interval = 5
	}
	duration := time.Duration(config.Interval)

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
