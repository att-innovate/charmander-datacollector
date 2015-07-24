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
	"fmt"
	"encoding/json"
)

type contextObj struct {
	Id int `json:"context"`
}

type ContextList struct {
	list map[string]int
}

func (instanceStore *ContextList) UpdateContext(hosts map[string]string) {

	fmt.Println("Updating pcp Context")

	for _, host := range hosts {

		content, err := getContent(fmt.Sprint("http://", host, ":44323/pmapi/context?hostspec=localhost&polltimeout=120"))
		if err != nil {
			fmt.Println("Cannot get context from:",host,".", err)
			continue
		}

		var context contextObj
		err = json.Unmarshal(content, &context)
		if err != nil {
			fmt.Println("Update Context json error:", err)

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
