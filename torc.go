package main

import (
	"encoding/json"
	"fmt"
)

type Nodes[] struct {
	Name string `json:"name"`
	Ip string `json:"ip"`
	Node_type string `json:"node_type"`
}

type TorcMeteredTask[] struct {
	Name string `json:"name"`
	Ip string `json:"ip"`
	Node_type string `json:"node_type"`
	Node_name string `json:"node_name"`
	Id string `json:"id"`
}

func getTorcNodes(url string) map[string]string {
	result := map[string]string{}
	content, err := getContent(fmt.Sprint("http://", url, "/nodes"))
	if err != nil {
		fmt.Println("Cannot get nodes from:", url, ".", err)
	}
	var nodes Nodes
	err = json.Unmarshal(content, &nodes)
	if err != nil {
		fmt.Println("Update node json error:", err)
	}
	for _, value :=  range nodes {
		if value.Node_type != "master" {
			result[value.Name]=value.Name
		}
	}

	return result
}

func TorcContainerMetered(containerName string, controller string) bool {
	content, err := getContent(fmt.Sprint("http://", controller, "/services/metered"))
	if err != nil {
		fmt.Println("Cannot get metered task from torc.", err)
	}
	var torcMeteredTask TorcMeteredTask
	err = json.Unmarshal(content, &torcMeteredTask)
	if err != nil {
		fmt.Println("Update node json error:", err)
	}

	for _, meteredContainerName := range torcMeteredTask {
		if meteredContainerName.Name == containerName {
			return true
		}
	}

	return false
}