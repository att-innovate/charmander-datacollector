package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

func meteredTask(host string, dockerId string) string{
	meteredTasks := make(map[string]string)
	var tempStr = fmt.Sprint("http://",host,":31300/getid/", dockerId)
	content, err := getContent(tempStr)
	taskName := strings.TrimSpace(string(content[:]))
	if err != nil {
		fmt.Println("error metered:", err)
	}

	meteredTasks[dockerId] = taskName

	if taskName, ok := meteredTasks[dockerId]; ok {
		if ContainerMetered(taskName){
			return taskName
		} else {
			return ""
		}
	} else {
		return ""
	}

}

func count(counter int) int{
	counter++
	return counter
}

func GetHost() {

	var hosts = GetCadvisorHosts()
	for _,host := range hosts {
		go collectData(host)
	}

}

func collectData(host string) {


		fmt.Println("host :",host)
		content, err := getContent(fmt.Sprint("http://",host,":44323/pmapi/context?hostspec=localhost&polltimeout=600"))
		if err != nil {
			fmt.Println("error:", err)
		}

		var context Context
		err2 := json.Unmarshal(content, &context)
		if err2 != nil {
			fmt.Println("error2:", err2)
		}

		var unmarshalledData Datametric

		response2, err3 := getContent(fmt.Sprint("http://",host,":44323/pmapi/", context.Id, DefaultStats))
		if err3 != nil {
			fmt.Println("error3:", err3)
		}

		err4 := json.Unmarshal(response2, &unmarshalledData)
		if err4 != nil {
			fmt.Println("error4:", err4)
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

		response, err5 := getContent(fmt.Sprint("http://",host,":44323/pmapi/", context.Id, secondCallParams))
		if err5 != nil {
			fmt.Println("error5:", err5)
		}
		err6 := json.Unmarshal(response, &secondCallData)
		if err6 != nil {
			fmt.Println("error6:", err6)
		}
		//fmt.Println(secondCallData)
		for _, b := range secondCallData.Instances {
			dataMap[unmarshalledData.Values[0].Name][b.Instance] = b.Name
		}

		for {
			c1 := make(chan []byte, 1)
			go func() {
				time.Sleep(time.Second * 2)
				c1 <- getData(host, context.Id, DefaultStats)
			}()

			select {
			case res := <-c1:
				processData(host, res, dataMap)
			case <-time.After(time.Second * 10):
				fmt.Println("timeout 1")
			}
		}


}

func getData(host string, context int, suffix string) []byte {
	var combinedURL = fmt.Sprint("http://",host,":44323/pmapi/", context, suffix)

	content, err := getContent(combinedURL)
	if err != nil {
		s := err.Error()
		return []byte(s)
	} else {
		return content
	}
}

func processData(host string, data []byte, instanceIdNameMapping map[string]map[int64]string) {

	var unmarshalledData2 Datametric

	err7 := json.Unmarshal(data, &unmarshalledData2)
	if err7 != nil {
		fmt.Println("error7:", err7)
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
			var cpuUsage = instance_name[0].Value

			var instance_name2 = filterByName(instanceOnly, "cgroup.memory.usage")
			var memoryUsage = instance_name2[0].Value

			machinePoints = append(machinePoints, []interface{}{
				instance_name[0].Timestamp,
				host, //hostname
				"/", //container_name
				memoryUsage, //memory usage
				nil,
				cpuUsage, //cpu_cumulative_usage
				nil,
				0, //rx bytes
				0, //rx error
				0, //tx bytes
				0, //tx error

			})
		}

		if !(strings.Contains(id, "docker")) || len(id) <8{
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
		var cpuUsage = instance_name2[0].Value

		statPoints = append(statPoints, []interface{}{
			instance_name[0].Timestamp,
			memoryUsage,                    //memory usage
			5983276,                        //page faults
			host,                       //hostname
			taskName, //container_name
			0,        //rxbyte
			0,        //rxerror
			0,        //txbyte
			0,        //txerrors
			cpuUsage, //cpu_cumulative_usage
			63733760, //memory_working_set

		})

	}

	if statPoints != nil {
		Write(statPoints, "stats")
		fmt.Println("hostname:",host)
		fmt.Println("wrote to stats db")
	}

	if machinePoints != nil {
		Write(machinePoints, "machine")
		fmt.Println("hostname:",host)
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
