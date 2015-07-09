package main

import (
	"sync"
)

type PreviousValue struct {
	Machine map[string]Metrics
	Stats map[string]Metrics
	sync.RWMutex
}

type Metrics struct {
	CPUUser      int64
	CPUSystem    int64
	MemoryUsage  int64
	MemorySystem int64
}

func NewValueStore() *PreviousValue {
	return &PreviousValue{
		Machine: make(map[string]Metrics),
		Stats: make(map[string]Metrics),
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

func (previousValue *PreviousValue) SearchById(s string) Metrics {
	previousValue.RLock()
	defer previousValue.RUnlock()
	return previousValue.Stats[s]
}

func (previousValue *PreviousValue) SearchByHost(host string) Metrics {
	previousValue.RLock()
	defer previousValue.RUnlock()
	return previousValue.Machine[host]
}