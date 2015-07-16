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
	"time"
	"flag"
	"fmt"
	"os"
)

var PcpMetrics = []string{
	"cgroup.cpuacct.stat.user",
	"cgroup.cpuacct.stat.system",
	"cgroup.memory.usage",
	"network.interface.in.bytes",
	"network.interface.out.bytes",
	"network.interface.out.drops",
	"network.interface.in.drops",
}

func main() {
	flag.Parse()

	var contextStore = NewContext()
	var hosts = GetCadvisorHosts()
	if len(hosts) == 0 {
		fmt.Println("Error: Could not talk to redis to obtain host")
		os.Exit(1)
	}
	contextStore.UpdateContext(hosts)

	GetInstanceMapping(contextStore)

	doWork(contextStore)
}

func doWork(contextStore *ContextList) {
	for host, _ := range contextStore.list {
		go func(host string, contextStore *ContextList) {
			for {
				var responseData = collectData(host, contextStore)
				if responseData.host != ""{
					processData(responseData)
				}
				time.Sleep(time.Second * 5)
			}
		}(host, contextStore)
	}

	keepAlive := make(chan int, 1)
	<-keepAlive
}