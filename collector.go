package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var PreviousValues = NewValueStore()
var instanceStore = NewInstanceStore()

type Param []struct {
	Instance []int
	Name     string
}

type Datametric struct {
	Timestamp struct {
		Time int64 `json:"s"`
		Us   int64 `json:"us"`
	}
	Values []struct {
		Pmid      int64  `json:"pmid"`
		Name      string `json:"name"`
		Instances []struct {
			Instance int64 `json:"instance"`
			Value    int64 `json:"value"`
		}
	}
}

type MetricsJsonStructure struct {
	Indom     int64 `json:"indom"`
	Instances []struct {
		Instance int64 `json:"instance"`
		Name     string `json:"name"`
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
	var requestURL = fmt.Sprint("http://", host, ":31300/getid/", dockerId)
	content, err := getContent(requestURL)
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

func GetInstanceMapping(context *ContextList) {
	instanceStore = NewInstanceStore()

	var hosts = GetCadvisorHosts()

	if len(hosts) == 0 {
		fmt.Println("Error: could not talk to redis to obtain host")
		os.Exit(1)
	}

	go func(context *ContextList) {
		for {

			hosts = GetCadvisorHosts()
			for _, host := range hosts {
				for _, value := range pcpMetrics {

					var requestMetricNames = fmt.Sprint("/_indom?&name=", value)

					var MetricsData MetricsJsonStructure

					response, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", context.list[host], requestMetricNames))
					if err != nil {
						fmt.Println("error5:", err)
					}
					err = json.Unmarshal(response, &MetricsData)
					if err != nil {
						fmt.Println("error6:", err)
					}

					for _, instance := range MetricsData.Instances {

						var instanceData = InstanceData{
							Host:       host,
							Instance:   instance.Instance,
							MetricName: value,
							Value:      instance.Name,
						}

						instanceStore.AddInstanceData(instanceData)
					}
				}
			}
			time.Sleep(time.Second * 30)
		}
	}(context)

}

func collectData(host string, contextStore *ContextList) GenericData {

	var combinedMetircString = strings.Join(pcpMetrics, ",")
	var unmarshalledData Datametric

	response, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", contextStore.list[host], "/_fetch?names=", combinedMetircString))
	if err != nil {
		fmt.Println("error3:", err)
	}

	err = json.Unmarshal(response, &unmarshalledData)
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

	var requestMetricNames = fmt.Sprint("/_indom?&name=", unmarshalledData.Values[0].Name)

	var MetricsData MetricsJsonStructure

	response, err = getContent(fmt.Sprint("http://", host, ":44323/pmapi/", contextStore.list[host], requestMetricNames))
	if err != nil {
		fmt.Println("error5:", err)
	}
	err = json.Unmarshal(response, &MetricsData)
	if err != nil {
		fmt.Println("error6:", err)
	}

	for _, b := range MetricsData.Instances {
		dataMap[unmarshalledData.Values[0].Name][b.Instance] = b.Name
	}

	dataFromGetData := getData(host, contextStore.list[host], fmt.Sprint("/_fetch?names=", combinedMetircString))

	var returnObj = GenericData{
		data:      dataFromGetData,
		datamap:   dataMap,
		contextid: contextStore.list[host],
		host:      host,
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
				Timestamp:  unmarshalledData2.Timestamp.Time,
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
	var networkPoints [][]interface{}

	for key := range instances {
		//fmt.Println("metrics:",metrics)
		var instanceOnly = filterByInstance(metrics, key)

		var instance_name = filterByName(instanceOnly, "cgroup.memory.usage")
		if len(instance_name) == 0 {
			continue
		}

		var instancestoreData = instanceStore.SearchByHost(host).SearchByMetric("network.interface.in.bytes").SearchByInstance(key)
		var interfaceName = ""

		if len(instancestoreData) != 0 {
			interfaceName = instancestoreData[0].Value

			var instance_name4 = filterByName(instanceOnly, "network.interface.in.bytes")
			var networkInBytes int64

			if len(instance_name4) != 0 {
				networkInBytes = instance_name4[0].Value
			}

			var instance_name5 = filterByName(instanceOnly, "network.interface.out.bytes")
			var networkOutBytes int64
			if len(instance_name5) != 0 {
				networkOutBytes = instance_name5[0].Value
			}

			var instance_name6 = filterByName(instanceOnly, "network.interface.out.drops")
			var networkInDrops int64
			if len(instance_name6) != 0 {
				networkInDrops = instance_name6[0].Value
			}

			var instance_name7 = filterByName(instanceOnly, "network.interface.in.drops")
			var networkOutDrops int64
			if len(instance_name7) != 0 {
				networkOutDrops = instance_name7[0].Value
			}

			if PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkInBytes == 0 {

				var metrics = Metrics{
					NetworkInBytes:  networkInBytes,
					NetworkOutBytes: networkOutBytes,
				}

				PreviousValues.AddNetworkMetrics(host, interfaceName, metrics)

			} else {

				var networkInBytesPoints = instance_name4[0].Value - PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkInBytes

				var networkOutBytesPoints = instance_name5[0].Value - PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkOutBytes

				networkPoints = append(networkPoints, []interface{}{

					instance_name[0].Timestamp,
					host, //hostname
					networkInBytesPoints,
					networkOutBytesPoints,
					interfaceName,
					networkInDrops,
					networkOutDrops,
				})

				var metrics = Metrics{
					NetworkInBytes:  networkInBytes,
					NetworkOutBytes: networkOutBytes,
				}
				PreviousValues.AddNetworkMetrics(host, interfaceName, metrics)
			}
		}

		if key == 0 {
			var instance_name = filterByName(instanceOnly, "cgroup.cpuacct.stat.system")
			var cpuUsageSystem = instance_name[0].Value

			var instance_name2 = filterByName(instanceOnly, "cgroup.memory.usage")
			var memoryUsage = instance_name2[0].Value

			var instance_name3 = filterByName(instanceOnly, "cgroup.cpuacct.stat.user")
			var cpuUsageUser = instance_name3[0].Value

			if PreviousValues.SearchByHost(host).CPUSystem == 0 {

				var metrics = Metrics{
					CPUSystem:   cpuUsageSystem,
					CPUUser:     cpuUsageUser,
					MemoryUsage: memoryUsage,
				}
				PreviousValues.AddMachineMetrics(host, metrics)

			} else {

				var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchByHost(host).CPUSystem) / float64(10.000)
				var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchByHost(host).CPUUser) / float64(10.000)

				machinePoints = append(machinePoints, []interface{}{
					instance_name[0].Timestamp,
					host,
					memoryUsage,
					CPUSystemPercentage,
					CPUUserPercentage,
				})

				var metrics = Metrics{
					CPUSystem:   cpuUsageSystem,
					CPUUser:     cpuUsageUser,
					MemoryUsage: memoryUsage,
				}
				PreviousValues.AddMachineMetrics(host, metrics)

			}

		}

		var id = instanceIdNameMapping["cgroup.memory.usage"][key]

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
				CPUSystem:   cpuUsageSystem,
				CPUUser:     cpuUsageUser,
				MemoryUsage: memoryUsage,
			}
			PreviousValues.AddStatsMetrics(taskName, metrics)

		} else {

			var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchById(taskName).CPUSystem) / float64(10.000)
			var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchById(taskName).CPUUser) / float64(10.000)

			statPoints = append(statPoints, []interface{}{
				instance_name[0].Timestamp,
				memoryUsage,
				host,
				taskName,
				CPUUserPercentage,
				CPUSystemPercentage,
			})

			var metrics = Metrics{
				CPUSystem:   cpuUsageSystem,
				CPUUser:     cpuUsageUser,
				MemoryUsage: memoryUsage,
			}
			PreviousValues.AddStatsMetrics(taskName, metrics)

		}

	}

	if statPoints != nil {
		Write(statPoints, "stats")
	}

	if machinePoints != nil {
		Write(machinePoints, "machine")
	}

	if networkPoints != nil {
		Write(networkPoints, "network")
	}
}

func filterByName(metrics []MetricModel, metricName string) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Metricname == metricName })
}

func filterByInstance(metrics []MetricModel, instanceId int64) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Instanceid == instanceId })
}

func filter(metrics []MetricModel, fn func(MetricModel) bool) []MetricModel {
	var results []MetricModel
	for _, value := range metrics {
		if fn(value) {
			results = append(results, value)
		}
	}
	return results
}