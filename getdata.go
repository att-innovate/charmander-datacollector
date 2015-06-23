package main

import (
	"io/ioutil"
	"net/http"
	"fmt"
	"strings"
	"encoding/json"
	"time"
	"./redis"//can commit this and line 118/119 out for testing
)

const Data1 = "/_fetch?names=pmcd.hostname"
const Data2 = "/_indom?instance=0,1,2,5,137,141,142,143&name=cgroup.cpuacct.stat.user"
const Data3 = "/_fetch?names=cgroup.cpuacct.stat.user,cgroup.cpuacct.stat.system,cgroup.memory.usage,network.interface.in.bytes,network.interface.out.bytes"

type Param []struct {
		  Instance []int
		  Name     string
	  }


type An struct {
	Context  int
}


type PmidMap struct {
	Pmid struct{
			 container map[int]string
		 }
}

type Datametric struct {
	Timestamp struct {
				  S int64
				  Us int64
			  }
	Values []struct{
		Pmid int64
		Name string
		Instances []struct{
			Instance int64
			Value int64
		}
	}
}

type DataMetric2 struct {
	Indom int64
	Instances []struct{
		Instance int64
		Name string
	}
}

/*type Timestamp struct {
	S int64
	US int64
}
type Values struct {
	Pmid int64
	Name string
	Instances []Instances
}
type Instances struct {
	Instance int64
	value int64
}
type Json1 struct {
	Timestamp Timestamp
	Values []Values
}*/

func getContent(url string) ([]byte, error) {
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return body, nil
}

func main(){

	//http://10.250.3.97:44323/pmapi/1980551490/_fetch?names=pmcd.hostname
	//{"timestamp":{"s":1434751576,"us":161856 }, "values":[{"pmid":8388629,"name":"pmcd.hostname","instances":[
	//{"instance":-1, "value":"bladerunner3" }]}]}

	//http://10.250.3.97:44323/pmapi/1980551490/_indom?instance=-1&name=pmcd.hostname
	//{"indom":4294967295,"instances":[]}

	//http://10.250.3.97:44323/pmapi/1980551490/_indom?instance=0,1,2,5,124,132,135&name=cgroup.cpuacct.stat.user

	/*{"indom":12582933,"instances":[
	{"instance":0,"name":"/" },
	{"instance":1,"name":"/user" },
	{"instance":2,"name":"/user/1000.user" },
	{"instance":5,"name":"/docker" },
	{"instance":124,"name":"/docker/2aa77a2806fb218b97f996818647980008088f0181eab79bd8456d6f94a6dc70" },
	{"instance":132,"name":"/docker/b23239a485b8d3d87a9489bd803d8319cdaef9642a1dfe4e35d62ede8cb690e3" },
	{"instance":135,"name":"/user/1000.user/184.session" }]}*/

	var zz = redis.GetCadvisorHosts()
	fmt.Println(zz)

	content, err := getContent("http://10.250.3.97:44323/pmapi/context?hostspec=localhost&polltimeout=600")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(content))
	}


	var an An
	err2 := json.Unmarshal(content, &an)
	if err != nil {
		fmt.Println("error:", err2)
	}

	var dm Datametric
	z3, err := getContent(fmt.Sprint("http://10.250.3.97:44323/pmapi/",an.Context,Data3))
	err3 := json.Unmarshal(z3, &dm)
	if err3 != nil {
		fmt.Println("error:", err3)
	}



	thisMap := make(map[string]map[int64]string)
	aMap := make(map[int64]string)
	//aMap["test"] = "yes"

	for a, _ := range dm.Values {
		thisMap[dm.Values[a].Name] = aMap

		for b, _ := range dm.Values[a].Instances{

			thisMap[dm.Values[a].Name][dm.Values[a].Instances[b].Instance] = ""
			//fmt.Println(dm.Values[a].Instances[b].Instance)
		}

	}

	fmt.Println(thisMap)

	fmt.Println("-----")
	//fmt.Println(getData(an.Context, Data3))

	//"/_indom?instance=0,1,2,5,137,141,142,143&name=cgroup.cpuacct.stat.user"

	var s = ""
	for a := range thisMap[dm.Values[0].Name]{
		s=fmt.Sprint(s,a,",")
		//fmt.Println(a)
	}
	s=strings.TrimSuffix(s,",")
	fmt.Println(s)

	var secondCallNum = s//"0,1,2,5,137,141,142,143"

	var secondCall = fmt.Sprint("/_indom?instance=",secondCallNum,"&name=",dm.Values[0].Name)
	fmt.Println("2-----")

	var dm2 DataMetric2

	z4, err := getContent(fmt.Sprint("http://10.250.3.97:44323/pmapi/",an.Context,secondCall))
	err4 := json.Unmarshal(z4, &dm2)
	if err4 != nil {
		fmt.Println("error:", err4)
	}

	fmt.Println(dm2)

	fmt.Println("3-----")

	for _,b := range dm2.Instances{
		thisMap[dm.Values[0].Name][b.Instance]=b.Name
	}

	fmt.Println(thisMap)


	fmt.Println("4-----")

	for {
		c1 := make(chan string, 1)
		go func() {
			time.Sleep(time.Second * 2)
			//getData(an.Context)
			c1 <- getData(an.Context, Data3)
		}()

		select {
		case res := <-c1:
			fmt.Println(res)
		case <-time.After(time.Second * 5):
			fmt.Println("timeout 1")
		}
	}



}

func getData(context int, suffix string)(result string){
	var zzz = fmt.Sprint("http://10.250.3.97:44323/pmapi/",context,suffix)

	//fmt.Printf(zzz)
	content2, err := getContent(zzz)
	if err != nil {
		s := err.Error()
		return s
	} else {
		return string(content2)
	}
}

func processMap(json string){




}
