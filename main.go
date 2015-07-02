package main

import "fmt"
//import "time"

var Machinedata = &Queue{nodes: make([]*Node, 3)}
var StatsData = &Queue{nodes: make([]*Node, 3)}


//0 	time
//1		sequence_number
//2		hostname
//3		container_name
//4		memory_usage
//5		page_faults
//6		cpu_cumulative_usage
//7		memory_working_set
//8		rx_bytes
//9		rx_errors
//10	tx_bytes
//11	tx_errors
func main() {
	//strs := []string{"first", "second"}
	//names := make([]interface{}, len(strs))
	//for i, s := range strs {
	//	names[i] = s
	//}

	fmt.Println("test1")
	Get()
	//Write()
	fmt.Println("test2")

}