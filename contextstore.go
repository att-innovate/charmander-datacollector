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
	"encoding/json"
	"time"
	"fmt"
)

type contextObj struct {
	Id int `json:"context"`
}

type ContextList struct {
	list map[string]int
}

func (instanceStore *ContextList) UpdateContext(hosts map[string]string) {

	unreachableHost := make(map[string]int)

	fmt.Println("Updating pcp Context")

	for _, host := range hosts {

		content, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/context?hostspec=localhost&polltimeout=120"))
		if err != nil {
			fmt.Println("Cannot get context from:", host, ".", err)
			if val, ok := unreachableHost[host]; ok {
			    unreachableHost[host]=val+1
			    continue
			} else if val > 5{
				break;
			}
		} else {

			var context contextObj
			err = json.Unmarshal(content, &context)
			if err != nil {
				fmt.Println("Update Context json error:", err)

			}
			instanceStore.addContext(host, context.Id)
		}
		instanceStore.retryHost(host)
	}
}

func NewContext() *ContextList {
	return &ContextList{list: make(map[string]int)}
}

func (contextList *ContextList) addContext(host string, context int) {
	contextList.list[host] = context
}

func (contextList *ContextList) Length() int {
	return len(contextList.list)
}

func (contextList *ContextList) retryHost(host string){
	go func(host string, contextList *ContextList) {
		for {
			time.Sleep(time.Second * 300)
			fmt.Println("Refreshing pcp context on: ", host)
			contextList.getContext(host)
		}
	}(host, contextList)
}

func (contextList *ContextList) getContext(host string){
	content, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/context?hostspec=localhost&polltimeout=120"))
	if err != nil {
		fmt.Println("Cannot get context from:", host, ".", err)
	}

	var context contextObj
	err = json.Unmarshal(content, &context)
	if err != nil {
		fmt.Println("GetContext json error:", err)

	}
	contextList.addContext(host, context.Id)
}