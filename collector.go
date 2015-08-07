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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var PreviousValues = NewValueStore()
var instanceStore = NewInstanceStore()

var statsPoints [][]interface{}
var machinePoints [][]interface{}
var networkPoints [][]interface{}

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
		Instance int64  `json:"instance"`
		Name     string `json:"name"`
	}
}

type MetricModel struct {
	Timestamp  int64
	Metricname string
	Instanceid int64
	Value      int64
	Host	   string
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
		fmt.Println("Error talking to metered service:", err)
		return ""
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

	go func(context *ContextList) {
		for {
			for host, contextId := range context.list {
				for _, value := range PcpMetrics {
					var requestMetricNames = fmt.Sprint("/_indom?&name=", value)
					var MetricsData MetricsJsonStructure

					response, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", contextId, requestMetricNames))
					if err != nil {
						fmt.Println("Failed fetching context from host. Error:", err)
						continue
					}
					err = json.Unmarshal(response, &MetricsData)
					if err != nil {
						fmt.Println("Failed unmarshalling context json. Error:", err)
						continue
					}

					for _, instance := range MetricsData.Instances {
						var instanceData = InstanceData{
							Host:       host,
							InstanceId: instance.Instance,
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
	var combinedMetircString = strings.Join(PcpMetrics, ",")
	var pcpMetric Datametric

	response, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/", contextStore.list[host], "/_fetch?names=", combinedMetircString))
	if err != nil {
		fmt.Println("Failed getting metrics from host. Error:", err)
		return GenericData{}
	}
	err = json.Unmarshal(response, &pcpMetric)
	if err != nil {
		fmt.Println("Failed unmarshalling metric json. Error:", err)
		return GenericData{}
	}

	dataMap := make(map[string]map[int64]string)
	instanceIdMap := make(map[int64]string)

	//initalize data structure with blank value
	for instanceId, _ := range pcpMetric.Values {
		dataMap[pcpMetric.Values[instanceId].Name] = instanceIdMap

		for instance, _ := range pcpMetric.Values[instanceId].Instances {
			dataMap[pcpMetric.Values[instanceId].Name][pcpMetric.Values[instanceId].Instances[instance].Instance] = ""
		}
	}

	var requestMetricNames = fmt.Sprint("/_indom?&name=", pcpMetric.Values[0].Name)
	var MetricsData MetricsJsonStructure

	//getting docker id from server
	response, err = getContent(fmt.Sprint("http://", host, ":44323/pmapi/", contextStore.list[host], requestMetricNames))
	if err != nil {
		fmt.Println("Failed getting docker id from host. Error:", err)
		return GenericData{}
	}
	err = json.Unmarshal(response, &MetricsData)
	if err != nil {
		fmt.Println("Failed unmarshalling docker id json. Error::", err)
		return GenericData{}
	}

	for _, instanceData := range MetricsData.Instances {
		dataMap[pcpMetric.Values[0].Name][instanceData.Instance] = instanceData.Name
	}
	dataFromGetData := getData(host, contextStore.list[host], fmt.Sprint("/_fetch?names=", combinedMetircString))

	return GenericData{
		data:      dataFromGetData,
		datamap:   dataMap,
		contextid: contextStore.list[host],
		host:      host,
	}
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

func processData(genericData GenericData) {
	var host = genericData.host
	var data = genericData.data
	var instanceIdNameMapping = genericData.datamap
	var unmarshalledData Datametric

	err := json.Unmarshal(data, &unmarshalledData)
	if err != nil {
		fmt.Println("Failed unmarshalling metric data json:", err)
		return
	}

	var metrics []MetricModel
	instances := make(map[int64]struct{})
	for instanceId, pcpMetrics := range unmarshalledData.Values {
		for _, metric := range unmarshalledData.Values[instanceId].Instances {
			var tempMetrics = MetricModel{
				Timestamp:  unmarshalledData.Timestamp.Time,
				Metricname: pcpMetrics.Name,
				Instanceid: metric.Instance,
				Value:      metric.Value,
				Host:		host,
			}
			metrics = append(metrics, tempMetrics)
			instances[metric.Instance] = struct{}{}
		}
	}

	for instanceId := range instances {
		var hostData = filterByHost(metrics, host)
		var instanceData = filterByInstance(hostData, instanceId)

		if instanceId == 0 {
			processMachineData(host, instanceData)
		}

		var instancestoreData = instanceStore.SearchByHost(host).SearchByMetric("network.interface.in.bytes").SearchByInstance(instanceId)
		//if len(instancestoreData) != 0 && len(instanceData) != 0 {
		if len(instancestoreData) != 0 {
			interfaceName := instancestoreData[0].Value
			processNetworkData(host, instanceData, interfaceName)
		}

		var taskName = getTaskName(host, instanceIdNameMapping["cgroup.memory.usage"][instanceId])

		if len(taskName) < 1{
			continue
		}

		processStatsData(host, instanceData, taskName)

	}

	if len(statsPoints) > 0 {
		var success = Write(statsPoints, "stats")
		if success {
			statsPoints = [][]interface{}{}
		}
	}

	if len(machinePoints) > 0 {
		var success = Write(machinePoints, "machine")
		if success {
			machinePoints = [][]interface{}{}
		}
	}

	if len(networkPoints) > 0 {
		var success = Write(networkPoints, "network")
		if success {
			networkPoints =[][]interface{}{}
		}
	}

	if len(machinePoints) > 100{
		clearPoints()
	}
}

func clearPoints(){
	statsPoints = [][]interface{}{}
	machinePoints = [][]interface{}{}
	networkPoints =[][]interface{}{}
}

func processNetworkData(host string, data []MetricModel, interfaceName string){

	var filteredData4 = filterByName(data, "network.interface.in.bytes")
	var networkInBytes int64
	if len(filteredData4) != 0 {
		networkInBytes = filteredData4[0].Value
	}

	var filteredData5 = filterByName(data, "network.interface.out.bytes")
	var networkOutBytes int64
	if len(filteredData5) != 0 {
		networkOutBytes = filteredData5[0].Value
	}

	var filteredData6 = filterByName(data, "network.interface.out.drops")
	var networkInDrops int64
	if len(filteredData6) != 0 {
		networkInDrops = filteredData6[0].Value
	}

	var filteredData7 = filterByName(data, "network.interface.in.drops")
	var networkOutDrops int64
	if len(filteredData7) != 0 {
		networkOutDrops = filteredData7[0].Value
	}

	if len(filteredData4) < 1 || len (filteredData5) < 1 {
		fmt.Println("Error: filteredData length is 0",filteredData4,filteredData5)
		return
	}

	if PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkInBytes == 0 {
		var metrics = Metrics{
			NetworkInBytes:  networkInBytes,
			NetworkOutBytes: networkOutBytes,
			TimeStamp: filteredData4[0].Timestamp,
		}
		PreviousValues.AddNetworkMetrics(host, interfaceName, metrics)
	} else {
		var timeDelta = filteredData4[0].Timestamp - PreviousValues.SearchByInterfaceHost(host, interfaceName).TimeStamp
		var networkInBytesPoints = (filteredData4[0].Value - PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkInBytes) / timeDelta
		var networkOutBytesPoints = (filteredData5[0].Value - PreviousValues.SearchByInterfaceHost(host, interfaceName).NetworkOutBytes) / timeDelta

		networkPoints = append(networkPoints, []interface{}{
			filteredData4[0].Timestamp,
			host,
			networkInBytesPoints,
			networkOutBytesPoints,
			interfaceName,
			networkInDrops,
			networkOutDrops,
		})

		var metrics = Metrics{
			NetworkInBytes:  networkInBytes,
			NetworkOutBytes: networkOutBytes,
			TimeStamp: filteredData4[0].Timestamp,
		}
		PreviousValues.AddNetworkMetrics(host, interfaceName, metrics)
	}
}

func processMachineData(host string, data []MetricModel){

	var filteredData = filterByName(data, "cgroup.cpuacct.stat.system")
	var cpuUsageSystem = filteredData[0].Value

	var filteredData2 = filterByName(data, "cgroup.memory.usage")
	var memoryUsage = filteredData2[0].Value

	var filteredData3 = filterByName(data, "cgroup.cpuacct.stat.user")
	var cpuUsageUser = filteredData3[0].Value

	if PreviousValues.SearchByHost(host).CPUSystem == 0 {
		var metrics = Metrics{
			CPUSystem:   cpuUsageSystem,
			CPUUser:     cpuUsageUser,
			MemoryUsage: memoryUsage,
			TimeStamp: filteredData[0].Timestamp,
		}
		PreviousValues.AddMachineMetrics(host, metrics)
	} else {
		var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchByHost(host).CPUSystem) / float64(10.000)
		var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchByHost(host).CPUUser) / float64(10.000)

		machinePoints = append(machinePoints, []interface{}{
			filteredData[0].Timestamp,
			host,
			memoryUsage,
			CPUSystemPercentage,
			CPUUserPercentage,
		})

		var metrics = Metrics{
			CPUSystem:   cpuUsageSystem,
			CPUUser:     cpuUsageUser,
			MemoryUsage: memoryUsage,
			TimeStamp: filteredData[0].Timestamp,
		}
		PreviousValues.AddMachineMetrics(host, metrics)
	}
}

func processStatsData(host string, data []MetricModel, taskName string){
	var filteredData = filterByName(data, "cgroup.memory.usage")
	var memoryUsage = filteredData[0].Value

	var filteredData2 = filterByName(data, "cgroup.cpuacct.stat.user")
	var cpuUsageUser = filteredData2[0].Value

	var filteredData3 = filterByName(data, "cgroup.cpuacct.stat.system")
	var cpuUsageSystem = filteredData3[0].Value

	if PreviousValues.SearchById(taskName).CPUSystem == 0 {

		var metrics = Metrics{
			CPUSystem:   cpuUsageSystem,
			CPUUser:     cpuUsageUser,
			MemoryUsage: memoryUsage,
			TimeStamp: filteredData2[0].Timestamp,
		}
		PreviousValues.AddStatsMetrics(taskName, metrics)

	} else {
		var CPUSystemPercentage = float64(cpuUsageSystem-PreviousValues.SearchById(taskName).CPUSystem) / float64(10.000)
		var CPUUserPercentage = float64(cpuUsageUser-PreviousValues.SearchById(taskName).CPUUser) / float64(10.000)

		statsPoints = append(statsPoints, []interface{}{
			filteredData2[0].Timestamp,
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
			TimeStamp: filteredData2[0].Timestamp,
		}
		PreviousValues.AddStatsMetrics(taskName, metrics)
	}
}

func getTaskName(host string, id string) string{
	if !(strings.Contains(id, "docker")) || len(id) < 8 {
		return ""
	}
	i := strings.LastIndex(id, "/")
	dockerId := id[i+1:]
	return meteredTask(host, dockerId)
}

func filterByName(metrics []MetricModel, metricName string) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Metricname == metricName })
}

func filterByInstance(metrics []MetricModel, instanceId int64) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Instanceid == instanceId })
}

func filterByHost(metrics []MetricModel, host string) []MetricModel {
	return filter(metrics, func(metric MetricModel) bool { return metric.Host == host })
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