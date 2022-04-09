[![Known Vulnerabilities](https://snyk.io/test/github/Sinnohd/fping/badge.svg)](https://snyk.io/test/github/Sinnohd/fping)

## infping Monitoring with fping/InfluxDB/Grafana + Daemon SystemD
Parse fping output, store result in influxdb 2.x and visualize with grafana.

#### Requirements:
##### Golang:
Install golang : https://golang.org/doc/install
##### Fping
```
$ sudo apt-get install fping
```
##### Influxdb
Having an Influxdb 2.x up and running, create a organisation and a bucket and create tokens for writing (infping) and reading (Grafana) to this bucket.

```
# Setup fresh InfluxDB installation and get bucket id
jb@mobihex:~/bin$ influx setup
> Welcome to InfluxDB 2.0!
? Please type your primary username jb
? Please type your password *********
? Please type your password again *********
? Please type your primary organization name acme
? Please type your primary bucket name infping
? Please type your retention period in hours, or 0 for infinite 0
? Setup with these parameters?
  Username:          jb
  Organization:      acme
  Bucket:            infping
  Retention Period:  infinite
 Yes
User	Organization	Bucket
jb	acme		infping

# get the bucket id, needed for token creation step
jb@mobihex:~/bin$ influx bucket list | grep infping
4f5dab66955ddf7d	infping		infinite	168h0m0s		785274777f4f03a4	implicit

# or just create bucket on an existing installation
jb@mobihex:~/bin$ influx bucket create -n infping -r 0
ID			Name	Retention	Shard group duration	Organization ID		Schema Type
4f5dab66955ddf7d	infping	infinite	168h0m0s		785274777f4f03a4	implicit

# Create auth tokens for write and read
jb@mobihex:~/bin$ influx auth create --org acme --write-bucket 4f5dab66955ddf7d
ID			Description	Token												User Name	User ID			Permissions
092aee8b39223000			<WRITE Token>	jb		092aedce4a623000	[write:orgs/785274777f4f03a4/buckets/4f5dab66955ddf7d]
jb@mobihex:~/bin$ influx auth create --org acme --read-bucket 4f5dab66955ddf7d
ID			Description	Token												User Name	User ID			Permissions
092af66d1be23000			<READ Token>	jb		092aedce4a623000	[read:orgs/785274777f4f03a4/buckets/4f5dab66955ddf7d]
```

#### Edit config.toml:
Check where fping binary is located.
If the service isn't able to create the logfile at the configured location, output will be re-directed to stdout.

```
[influxdb]

host = "localhost"
port = "8086"
org = "acme"
bucket = "fping"
token = "<Influx v2 write auth token>"
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

##### Systemd

```
$ ./setup.sh
$ sudo systemctl status infping.service

```
##### Docker container

```
$ cp config.toml.example config.toml
# Adjust config.toml file
$ vi config.toml
# Create binary
$ go mod tidy
$ CGO_ENABLED=0 go build
# Build container
$ docker build . -t sinnohd/infping:0.2.0
$ docker run sinnohd/infping:0.2.0

```


#### Output
```
Feb 24 15:14:30 ip-172-19-64-10 infping: 2021/02/24 15:14:30 Connected to influxdb! (dur:9.877542ms, ver:1.8.0)
Feb 24 15:14:30 ip-172-19-64-10 infping: 2021/02/24 15:14:30 Going to ping the following ips: [192.168.0.1 192.168.0.2]
Feb 24 15:14:40 ip-172-19-64-10 infping: 2021/02/24 15:14:40 IP:192.168.0.1, send:10, recv:10 loss: 0, min: 1.95, avg: 2.13, max: 2.70
Feb 24 15:14:40 ip-172-19-64-10 infping: 2021/02/24 15:14:40 IP:192.168.0.2, send:10, recv:10 loss: 0, min: 289, avg: 289, max: 291
```

#### Todo


