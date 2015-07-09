package main

import (
	"fmt"
	"encoding/json"
)

type contextObj struct {
	Id int `json:"context"`
}

type ContextList struct {
	list map[string]int
}
func (instanceStore *ContextList) UpdateContext() {

	var hosts = GetCadvisorHosts()

	for _, host := range hosts {

		content, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/context?hostspec=localhost&polltimeout=600"))
		if err != nil {
			fmt.Println("error:", err)
		}

		var context contextObj
		err = json.Unmarshal(content, &context)
		if err != nil {
			fmt.Println("error2:", err)
		}

		instanceStore.addContext(host, context.Id)

	}

}

func NewContext() *ContextList{
	return &ContextList{list:make(map[string]int)}
}

func (contextList *ContextList) addContext(host string, context int) {
	contextList.list[host] = context
}
