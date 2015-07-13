package main
import (
	"strings"
	"time"
	"fmt"
	"runtime"
)

type Settings struct {
	Metrics []string
}

const DefaultStats = "cgroup.cpuacct.stat.user,cgroup.cpuacct.stat.system,cgroup.memory.usage,network.interface.in.bytes,network.interface.out.bytes,network.interface.out.drops,network.interface.in.drops"

var pcpMetrics []string

func main() {

	pcpMetrics = strings.Split(DefaultStats, ",")

	var contextStore = NewContext()

	contextStore.UpdateContext()

	GetInstanceMapping(contextStore)

	doWork(contextStore)

}

func doWork(contextStore *ContextList) {

	var hosts = GetCadvisorHosts()
	c1 := make(chan GenericData, 1)

	for {

		for _, host := range hosts {

			go func(host string) {

				c1 <- collectData(host,contextStore)

			}(host)

		}

		time.Sleep(time.Second * 5)

		for i := 0; i < len(hosts); i++ {

			res := <-c1

			processData(res)
			fmt.Println("GoRoutines:", runtime.NumGoroutine())

		}

	}
}