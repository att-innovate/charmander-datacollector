package main

type InstanceData struct {
	Host string
	InstanceId int64
	MetricName string
	Value string
}

type InstanceStore []InstanceData

func NewInstanceStore() *InstanceStore {
	return &InstanceStore{}
}

func (instanceStore *InstanceStore) AddInstanceData(instanceData InstanceData) {
	*instanceStore = append(*instanceStore, instanceData)
}

func (instanceStore InstanceStore) SearchByHost(host string) InstanceStore {
	return instancefilterByHost(instanceStore, host)
}

func (instanceStore InstanceStore) SearchByMetric(metric string) InstanceStore {
	return instancefilterByMetric(instanceStore, metric)
}

func (instanceStore InstanceStore) SearchByInstance(instanceId int64) InstanceStore {
	return instancefilterByInstanceId(instanceStore, instanceId)
}

func instancefilterByHost(instancedata InstanceStore, host string) InstanceStore {
	return instanceFilter(instancedata, func(metric InstanceData) bool { return metric.Host == host })
}

func instancefilterByMetric(instancedata InstanceStore, metricName string) InstanceStore {
	return instanceFilter(instancedata, func(metric InstanceData) bool { return metric.MetricName == metricName })
}

func instancefilterByInstanceId(instancedata InstanceStore, instanceId int64) InstanceStore {
	return instanceFilter(instancedata, func(metric InstanceData) bool { return metric.InstanceId == instanceId })
}

func instanceFilter(instancestore InstanceStore, fn func(InstanceData) bool) InstanceStore {
	var results InstanceStore
	for _, value := range instancestore {
		if fn(value) {
			results = append(results, value)
		}
	}
	return results
}
