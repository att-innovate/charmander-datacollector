package main
//package main
import (
	"github.com/influxdb/influxdb/client"
	"flag"
	"fmt"
	//"net/url"
	//"log"
	//"strconv"
	"time"
	//"math/rand"
	//"./influxdb"
)
//http://172.31.2.11:31400/
var Machine = []string{"time","hostname","container_name","memory_usage","page_faults","cpu_cumulative_usage","memory_working_set","rx_bytes","rx_errors","tx_bytes","tx_errors"}
var Stats = []string{"time",
	"memory_usage",
	"page_faults",
	"hostname",
	"container_name",
	"rx_bytes",
	"rx_errors",
	"tx_bytes",
	"tx_errors",
	"cpu_cumulative_usage",
	"memory_working_set"}

var (
	argBufferDuration = flag.Duration("sink_influxdb_buffer_duration", 10*time.Second, "Time duration for which stats should be buffered in influxdb sink before being written as a single transaction")
	argDbUsername     = flag.String("sink_influxdb_username", "root", "InfluxDB username")
	argDbPassword     = flag.String("sink_influxdb_password", "root", "InfluxDB password")
	argDbHost         = flag.String("sink_influxdb_host", "172.31.2.11:31410", "InfluxDB host:port")
	argDbName         = flag.String("sink_influxdb_name", "charmander", "Influxdb database name")
)

//func main() {
//	internalTest()
//}

func Write(data [][]interface{}, dataType string) {

	c, err := client.NewClient(&client.ClientConfig{
		Host:     *argDbHost,
		Username: *argDbUsername,
		Password: *argDbPassword,
		Database: *argDbName,
	})


	c.DisableCompression()

	if err != nil {
		panic(err)
	}

	//var	name = "ts9_uncompressed"

	result, err := c.Query("select * from machine")
	fmt.Println(result[0].Name)

	//for {
		a := StatsData.Pop()
		if a != nil {
			series := &client.Series{
				Name:    "machine",
				Columns: Machine,
				Points: data,
/*				[][]interface{}{
					{
						(time.Now().UnixNano()/1000000)/1000,
						"slave5", //hostname
						"/", //container_name
						532992000, //memory usage
						nil,
						29316732528, //cpu_cumulative_usage
						nil,
						67168, //rx bytes
						0, //rx error
						57745, //tx bytes
						0, //tx error
					},
				},*/
			}


			if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
				fmt.Println("failed to write stats to influxDb - %s", err)
			}
		}
		//b := StatsData.Pop()
		//if b != nil {

		if dataType == "stats"{
			//fmt.Println(b.Value)
			series := &client.Series{
				Name:    "stats",
				Columns: Stats,
				Points: data,

/*				[][]interface{}{
					{
						(time.Now().UnixNano()/1000000)/1000,
						63746048,//memory usage
						5983276,//page faults
						"slave5",//hostname
						"stress60-1435691470902714993",//container_name
						0,//rxbyte
						0,//rxerror
						0,//txbyte
						0,//txerrors
						66058462440,//cpu_cumulative_usage
						63733760,//memory_working_set
					},
				},*/
			}


			if err := c.WriteSeriesWithTimePrecision([]*client.Series{series}, client.Second); err != nil {
				fmt.Println("failed to write stats to influxDb - %s", err)
			}

			fmt.Println("wrote to db")
		}


	//}
}