# Datacollector
---
Datacollector enables Container Cluster Monitoring. 

Internally, datacollector uses [PCP](https://github.com/performancecopilot/pcp) 3.10.5+ for compute resource usage metrics.

Datacollector currently supports [Charmander](https://github.com/att-innovate/charmander). 


## Running Datacollector on Charmander


Datacollector grabs host data from redis then writes the data to InfluxDB.
It only support [InfluxDB](http://influxdb.com) 0.8x.

There are 3 time series created from this machine, stats and network. The stats time series will only be created if there are metered tasks in redis.

## Starting Datacollector

to be updated.