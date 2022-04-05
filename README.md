[![Known Vulnerabilities](https://snyk.io/test/github/Sinnohd/fping/badge.svg)](https://snyk.io/test/github/Sinnohd/fping)

## infping Monitoring with fping/InfluxDB/Grafana + Daemon SystemD
Parse fping output, store result in influxdb 1.x, and visualize with grafana.


#### Requirement:
##### Golang:
Install golang : https://golang.org/doc/install
##### Fping
```
$ sudo apt-get install fping
```

#### Edit config.toml:

```
[influxdb]

host = "localhost"
port = "8086"
org = "acme"
bucket = "fping"
measurement = "ping"
precision = "ms"
retentionpolicy = "infinite"
token = "<Influx v2 Auth Token>"
fping = "/usr/bin/fping"

[logs]
logfile = "/var/log/infping/infping.log"

[hosts]

hosts = [
    "192.168.0.1",
    "192.168.0.2",
]

```
#### Install infping:
```
$ ./setup.sh
$ sudo systemctl status infping.service

```

#### Output
```
Feb 24 15:14:30 ip-172-19-64-10 infping: 2021/02/24 15:14:30 Connected to influxdb! (dur:9.877542ms, ver:1.8.0)
Feb 24 15:14:30 ip-172-19-64-10 infping: 2021/02/24 15:14:30 Going to ping the following ips: [192.168.0.1 192.168.0.2]
Feb 24 15:14:40 ip-172-19-64-10 infping: 2021/02/24 15:14:40 IP:192.168.0.1, send:10, recv:10 loss: 0, min: 1.95, avg: 2.13, max: 2.70
Feb 24 15:14:40 ip-172-19-64-10 infping: 2021/02/24 15:14:40 IP:192.168.0.2, send:10, recv:10 loss: 0, min: 289, avg: 289, max: 291
```

#### Todo
Replace clientv2 for InfluxDB 1.x with InfluxDB 2 Go Client 

