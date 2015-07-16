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
