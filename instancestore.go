package main


import (
"sync"
)


func GetInstanceMapping(metricsName string, host string, context strub) PmidMap{

	var instanceMap = make(MetricName)
	var metricNameMap = make(PmidMap)

	var metricNameArray = strings.Split(DefaultStats, ",")

	for _, value := range metricNameArray{
		metricNameMap[value]=instanceMap

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

	}

	//fmt.Println(metricNameMap)

	return metricNameMap
}