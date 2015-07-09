package main

import "testing"



func TestMain(t *testing.T){

	var test = NewInstanceStore()

	var instanceData = InstanceData{
		Host:"test1",
		Instance:1,
		MetricName:"stat",
		Value:"docker1",
	}

	test.AddInstanceData(instanceData)

	var instanceData2 = InstanceData{
		Host:"test3",
		Instance:2,
		MetricName:"stat",
		Value:"docker2",
	}
	test.AddInstanceData(instanceData2)

	exists := test.SearchByHost("test3")
	if exists[0] != instanceData2 {
		t.Error("expected the item to match")
	}
	exists = test.SearchByMetric("stat")
	if len(exists) != 2 {
		t.Error("expected the item to match")
	}
	exists = test.SearchByInstance(1)
	if exists[0] != instanceData {
		t.Error("expected the item to match")
	}

}