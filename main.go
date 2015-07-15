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