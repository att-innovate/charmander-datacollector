package main

import (
	"flag"
	"bufio"
	"net"
	"strconv"
	"time"

	"github.com/golang/glog"

)
var redisHost = flag.String("source_redis_host", "127.0.0.1:6379", "Redis IP Address:Port")


func ContainerReady(containerName string) bool {
	if redis := redisAvailable(); redis != nil {
		defer redis.Close()
		sendCommand(redis, "KEYS", "charmander:tasks:*")
		containersReady := *parseResult(redis, "charmander:tasks:")
		for _, containerReady := range containersReady {
			if containerReady == containerName {
				return true
			}
		}
	}

	return false
}

func ContainerMetered(containerName string) bool {
	if redis := redisAvailable(); redis != nil {
		defer redis.Close()
		sendCommand(redis, "KEYS", "charmander:tasks-metered:*")
		meteredContainerNames := *parseResult(redis, "charmander:tasks-metered:")
		for _, meteredContainerName := range meteredContainerNames {
			if meteredContainerName == containerName {
				return true
			}
		}
	}

	return false
}

func GetCadvisorHosts() (*map[string]string) {
	result := map[string]string{}

	if redis := redisAvailable(); redis != nil {
		sendCommand(redis, "KEYS", "charmander:nodes:*")
		hosts := *parseResult(redis, "charmander:nodes:")
		for _, host := range hosts {
			result[host] = host
		}
		redis.Close()
	}

	return &result
}

func redisAvailable() net.Conn {
	connection, error := net.DialTimeout("tcp", *redisHost, 2 * time.Second)
	if error != nil {
		return nil
	}

	return connection
}

func sendCommand(connection net.Conn, args ...string) {
	buffer := make([]byte, 0, 0)
	buffer = encodeReq(buffer, args)
	connection.Write(buffer)
}

func parseResult(connection net.Conn, prefix string) *[]string {
	bufferedInput := bufio.NewReader(connection)
	line, _, err := bufferedInput.ReadLine()
	if err != nil {
		glog.Errorf("error parsing redis response %s\n", err)
		return &[]string {}
	}
	numberOfArgs, _ := strconv.ParseInt(string(line[1:]), 10, 64)
	args := make([]string, 0, numberOfArgs)
	for i := int64(0); i < numberOfArgs; i++ {
		line, _, _ = bufferedInput.ReadLine()
		argLen, _ := strconv.ParseInt(string(line[1:]), 10, 32)
		line, _, _ = bufferedInput.ReadLine()
		args = append(args, string(line[len(prefix):argLen]))
	}

	return &args
}

func encodeReq(buf []byte, args []string) []byte {
	buf = append(buf, '*')
	buf = strconv.AppendUint(buf, uint64(len(args)), 10)
	buf = append(buf, '\r', '\n')
	for _, arg := range args {
		buf = append(buf, '$')
		buf = strconv.AppendUint(buf, uint64(len(arg)), 10)
		buf = append(buf, '\r', '\n')
		buf = append(buf, []byte(arg)...)
		buf = append(buf, '\r', '\n')
	}
	return buf
}
