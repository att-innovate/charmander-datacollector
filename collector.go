package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var PreviousValues = NewValueStore()

const DefaultStats = "/_fetch?names=cgroup.cpuacct.stat.user,cgroup.cpuacct.stat.system,cgroup.memory.usage,network.interface.in.bytes,network.interface.out.bytes"

type Param []struct {
	Instance []int
	Name     string
}

type Context struct {
	Id int `json:"context"`
}

type PmidMap struct {
	Pmid struct {
		container map[int]string
	}
}
//todo: rename object
type Datametric struct {
	Timestamp struct {
		S  int64
		Us int64
	}
	Values []struct {
		Pmid      int64
		Name      string
		Instances []struct {
			Instance int64
			Value    int64
		}
	}
}

type SecondCallDataMetric struct {
	Indom     int64
	Instances []struct {
		Instance int64
		Name     string
	}
}

type MetricModel struct {
	Timestamp  int64
	Metricname string
	Instanceid int64
	Value      int64
}

type GenericData struct {
	host      string
	contextid int
	datamap   map[string]map[int64]string
	data      []byte
}

func getContent(url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func meteredTask(host string, dockerId string) string {
	meteredTasks := make(map[string]string)
	var tempStr = fmt.Sprint("http://", host, ":31300/getid/", dockerId)
	content, err := getContent(tempStr)
	taskName := strings.TrimSpace(string(content[:]))
	if err != nil {
		fmt.Println("error metered:", err)
	}

	meteredTasks[dockerId] = taskName

	if taskName, ok := meteredTasks[dockerId]; ok {
		if ContainerMetered(taskName) {
			return taskName
		} else {
			return ""
		}
	} else {
		return ""
	}

}

func GetHost() {

	var hosts = GetCadvisorHosts()
	c1 := make(chan GenericData, 1)

	for {

		for _, host := range hosts {

			go func(host string) {

				c1 <- collectData(host)

			}(host)

		}
		time.Sleep(time.Second * 10)
		for i := 0; i < len(hosts); i++ {
			res := <-c1

			processData(res)
			fmt.Println("GoRoutines:", runtime.NumGoroutine())
			fmt.Println("NumCgoCall:", runtime.NumCgoCall())
		}

	}
}

func collectData(host string) GenericData {

	content, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/context?hostspec=localhost&polltimeout=600"))
	if err != nil {
		fmt.Println("error:", err)
	}

	var context Context
	err = json.Unmarshal(content, &context)
	if err != nil {
		fmt.Println("error2:", err)
	}

	var unmarshalledData Datametric

	response2, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", context.Id, DefaultStats))
	if err != nil {
		fmt.Println("error3:", err)
	}

	err = json.Unmarshal(response2, &unmarshalledData)
	if err != nil {
		fmt.Println("error4:", err)
	}

	dataMap := make(map[string]map[int64]string)
	instanceIdMap := make(map[int64]string)

	for a, _ := range unmarshalledData.Values {
		dataMap[unmarshalledData.Values[a].Name] = instanceIdMap

		for b, _ := range unmarshalledData.Values[a].Instances {

			dataMap[unmarshalledData.Values[a].Name][unmarshalledData.Values[a].Instances[b].Instance] = ""

		}
	}

	var s = ""
	for a := range dataMap[unmarshalledData.Values[0].Name] {
		s = fmt.Sprint(s, a, ",")
	}
	s = strings.TrimSuffix(s, ",")

	var instanceNum = s

	var secondCallParams = fmt.Sprint("/_indom?instance=", instanceNum, "&name=", unmarshalledData.Values[0].Name)

	var secondCallData SecondCallDataMetric

	response, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", context.Id, secondCallParams))
	if err != nil {
		fmt.Println("error5:", err)
	}
	err = json.Unmarshal(response, &secondCallData)
	if err != nil {
		fmt.Println("error6:", err)
	}
	//fmt.Println(secondCallData)
	for _, b := range secondCallData.Instances {
		dataMap[unmarshalledData.Values[0].Name][b.Instance] = b.Name
	}

	dataFromGetData := getData(host, context.Id, DefaultStats)

	var returnObj = GenericData{
		data:      dataFromGetData,
		datamap:   dataMap,
		contextid: context.Id,
		host:      host,
		//PreviousData:map[string]PreviousValue{},
	}
	return returnObj

}

func getData(host string, context int, suffix string) []byte {
	var combinedURL = fmt.Sprint("http://", host, ":44323/pmapi/", context, suffix)

	content, err := getContent(combinedURL)
	if err != nil {
		s := err.Error()
		return []byte(s)
	} else {
		return content
	}
}

//func processData(host string, data []byte, instanceIdNameMapping map[string]map[int64]string) {
func processData(gData GenericData) {
	var host = gData.host
	var data = gData.data
	var instanceIdNameMapping = gData.datamap

	var unmarshalledData2 Datametric

	err := json.Unmarshal(data, &unmarshalledData2)
	if err != nil {
		fmt.Println("error7:", err)
	}

	var metrics []MetricModel

	instances := make(map[int64]struct{})

	for a, b := range unmarshalledData2.Values {
		for _, c := range unmarshalledData2.Values[a].Instances {
			var tempMetrics = MetricModel{
				Timestamp:  unmarshalledData2.Timestamp.S,
				Metricname: b.Name,
				Instanceid: c.Instance,
				Value:      c.Value,
			}
			metrics = append(metrics, tempMetrics)
			instances[c.Instance] = struct{}{}

		}
	}

	var statPoints [][]interface{}
	var machinePoints [][]interface{}

	for key := range instances {

		var instanceOnly = filterByInstance(metrics, key)
		var instance_name = filterByName(instanceOnly, "cgroup.memory.usage")
		if len(instance_name) == 0 {
			continue
		}

		var id = instanceIdNameMapping["cgroup.memory.usage"][key]

		if key == 0 {
			var instance_name = filterByName(instanceOnly, "cgroup.cpuacct.stat.system")
			var cpuUsageSystem = instance_name[0].Value

			var instance_name2 = filterByName(instanceOnly, "cgroup.memory.usage")
			var memoryUsage = instance_name2[0].Value

			var instance_name3 = filterByName(instanceOnly, "cgroup.cpuacct.stat.user")
			var cpuUsageUser = instance_name3[0].Value

			fmt.Println("CPUSystem:", cpuUsageSystem)
			fmt.Println("CPUUser:", cpuUsageUser)

			if PreviousValues.SearchByHost(host).CPUSystem == 0 {

				var metrics = Metrics{
					CPUSystem: cpuUsageSystem,
					CPUUser : cpuUsageUser,
					MemoryUsage : memoryUsage,
				}
				PreviousValues.AddMachineMetrics(host, metrics)


			} else {

				var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchByHost(host).CPUSystem) / float64(10.000)
				var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchByHost(host).CPUUser) / float64(10.000)

				//"kernel.all.cpu.sys",
				//"kernel.all.cpu.user"

				machinePoints = append(machinePoints, []interface{}{

					instance_name[0].Timestamp,
					host,        //hostname
					memoryUsage, //memory usage
					nil,
					CPUSystemPercentage, //cpu_cumulative_usage
					nil,
					CPUUserPercentage, //cpuUsageUser
				})

				var metrics = Metrics{
					CPUSystem: cpuUsageSystem,
					CPUUser : cpuUsageUser,
					MemoryUsage : memoryUsage,
				}
				PreviousValues.AddMachineMetrics(host, metrics)

			}

		}

		if !(strings.Contains(id, "docker")) || len(id) < 8 {
			continue
		}

		i := strings.LastIndex(id, "/")

		dockerId := id[i+1:]
		var taskName string

		taskName = meteredTask(host, dockerId)

		if taskName == "" {
			continue
		}

		var memoryUsage = instance_name[0].Value

		var instance_name2 = filterByName(instanceOnly, "cgroup.cpuacct.stat.user")
		var cpuUsageUser = instance_name2[0].Value

		var instance_name3 = filterByName(instanceOnly, "cgroup.cpuacct.stat.system")
		var cpuUsageSystem = instance_name3[0].Value

		if PreviousValues.SearchById(taskName).CPUSystem == 0 {

			var metrics = Metrics{
				CPUSystem: cpuUsageSystem,
				CPUUser : cpuUsageUser,
				MemoryUsage : memoryUsage,
			}
			PreviousValues.AddStatsMetrics(taskName, metrics)

		} else {

			var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchById(taskName).CPUSystem) / float64(10.000)
			var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchById(taskName).CPUUser) / float64(10.000)


			statPoints = append(statPoints, []interface{}{
				instance_name[0].Timestamp,
				memoryUsage,         //memory usage
				5983276,             //page faults
				host,                //hostname
				taskName,            //container_name
				CPUUserPercentage,   //cpu_cumulative_usage
				63733760,            //memory_working_set
				CPUSystemPercentage, //cpuUsageSystem

			})

			var metrics = Metrics{
				CPUSystem: cpuUsageSystem,
				CPUUser : cpuUsageUser,
				MemoryUsage : memoryUsage,
			}
			PreviousValues.AddStatsMetrics(taskName, metrics)

		}

	}

	if statPoints != nil {
		Write(statPoints, "stats")
		fmt.Println("hostname:", host)
		fmt.Println("wrote to stats db")
	}

	if machinePoints != nil {
		Write(machinePoints, "machine")
		fmt.Println("hostname:", host)
		fmt.Println("wrote to machine db")
	}
}

func filterByName(metrics []MetricModel, metricName string) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Metricname == metricName })
}

func filterByInstance(metrics []MetricModel, instanceId int64) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Instanceid == instanceId })
}

func filter(s []MetricModel, fn func(MetricModel) bool) []MetricModel {
	var r []MetricModel
	for _, v := range s {
		if fn(v) {
			r = append(r, v)
		}
	}
	return r
}
