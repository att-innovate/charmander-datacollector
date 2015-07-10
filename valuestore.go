package main

import (
	"sync"
)

type PreviousValue struct {
	Machine map[string]Metrics
	Stats map[string]Metrics
	Network map[string]Metrics
	sync.RWMutex
}

type Metrics struct {
	CPUUser      int64
	CPUSystem    int64
	MemoryUsage  int64
	MemorySystem int64
	NetworkInBytes int64
	NetworkOutBytes int64
}

func NewValueStore() *PreviousValue {
	return &PreviousValue{
		Machine: make(map[string]Metrics),
		Stats: make(map[string]Metrics),
		Network: make(map[string]Metrics),
	}
}

func (previousValue *PreviousValue) AddMachineMetrics(host string, metric Metrics) {
	previousValue.Lock()
	defer previousValue.Unlock()
	previousValue.Machine[host] = metric
}

func (previousValue *PreviousValue) AddStatsMetrics(containername string, metric Metrics) {
	previousValue.Lock()
	defer previousValue.Unlock()
	previousValue.Stats[containername] = metric
}

func (previousValue *PreviousValue) AddNetworkMetrics(host string, interfacename string, metric Metrics) {
	previousValue.Lock()
	defer previousValue.Unlock()
	previousValue.Network[host+interfacename] = metric
}

func (previousValue *PreviousValue) SearchById(id string) Metrics {
	previousValue.RLock()
	defer previousValue.RUnlock()
	return previousValue.Stats[id]
}

func (previousValue *PreviousValue) SearchByHost(host string) Metrics {
	previousValue.RLock()
	defer previousValue.RUnlock()
	return previousValue.Machine[host]
}

func (previousValue *PreviousValue) SearchByInterfaceHost(host string, interfacename string) Metrics {
	previousValue.RLock()
	defer previousValue.RUnlock()
	return previousValue.Network[host+interfacename]
}