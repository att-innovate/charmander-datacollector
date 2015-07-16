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