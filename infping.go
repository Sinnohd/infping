package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	toml "github.com/pelletier/go-toml/v2"
)

const (
	path = "config.toml"
)

// Hard error, log & exit (equivalent to log.Fatal)
func herr(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// Soft error, print out and go forward
func perr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func slashSplitter(c rune) bool {
	return c == '/'
}

func readPoints(config InfConfig, client influxdb2.Client, logger *log.Logger) {
	nodes := config.hosts.hosts
	args := []string{"-B 1", "-D", "-r0", "-O 0", "-Q 10", "-p 1000", "-l"}
	list := []string{}
	for _, node := range nodes {
		args = append(args, node)
		list = append(list, node)
	}
	logger.Printf("Going to ping the following ips: %v", list)
	cmd := exec.Command(config.influxdb.fping, args...)

	stdout, err := cmd.StdoutPipe()
	herr(err)
	stderr, err := cmd.StderrPipe()
	herr(err)
	err = cmd.Start()
	perr(err)

	buff := bufio.NewScanner(stderr)
	for buff.Scan() {
		text := buff.Text()
		fields := strings.Fields(text)
		// Ignore timestamp
		if len(fields) > 1 {
			ip := fields[0]
			data := fields[4]
			dataSplitted := strings.FieldsFunc(data, slashSplitter)
			// Remove ,
			dataSplitted[2] = strings.TrimRight(dataSplitted[2], "%,")
			sent, recv, lossp := dataSplitted[0], dataSplitted[1], dataSplitted[2]
			min, max, avg := "", "", ""
			// Ping times
			if len(fields) > 5 {
				times := fields[7]
				td := strings.FieldsFunc(times, slashSplitter)
				min, avg, max = td[0], td[1], td[2]
			}
			logger.Printf("IP:%s, send:%s, recv:%s loss: %s, min: %s, avg: %s, max: %s", ip, sent, recv, lossp, min, avg, max)
			writePoints(config, logger, client, ip, sent, recv, lossp, min, avg, max)
		}

	}
	std := bufio.NewReader(stdout)
	line, err := std.ReadString('\n')
	perr(err)
	logger.Printf("stdout:%s", line)
}

func writePoints(config InfConfig, logger *log.Logger, client influxdb2.Client, ip string, sent string, recv string, lossp string, min string, avg string, max string) {
	ms := config.influxdb.measurement
	org := config.influxdb.org
	bucket := config.influxdb.bucket

	loss, _ := strconv.Atoi(lossp)
	fields := map[string]interface{}{}
	if min != "" && avg != "" && max != "" {
		min, _ := strconv.ParseFloat(min, 64)
		avg, _ := strconv.ParseFloat(avg, 64)
		max, _ := strconv.ParseFloat(max, 64)
		fields = map[string]interface{}{
			"loss": loss,
			"min":  min,
			"avg":  avg,
			"max":  max,
		}
	} else {
		fields = map[string]interface{}{
			"loss": loss,
		}
	}

	// Create a point and add to batch
	tags := map[string]string{
		"addr": ip,
	}

	writeApi := client.WriteAPI(org, bucket)
	// create point
	p := influxdb2.NewPoint(ms, tags, fields, time.Now())
	writeApi.WritePoint(p)
	writeApi.Flush()
}

// begin config file structs
type InfConfig struct {
	influxdb influxdb
	logs logs
	hosts hosts
}

type influxdb struct {
	host string
	port string
	org string
	bucket string
	measurement string
	token string
	fping string
}

type logs struct {
	logfile string
}

type hosts struct {
	hosts []string
}
// end config file structs

func main() {
	var config InfConfig
	file, err := os.ReadFile(path)
	herr(err)
	err = toml.Unmarshal(file, &config)
	herr(err)

	host := config.influxdb.host
	port := config.influxdb.port
	token := config.influxdb.token

	logfile := config.logs.logfile
	logger := log.New(os.Stdout, "", log.LstdFlags)

	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Println(err)
	} else {
		logger.SetOutput(f)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.Printf("Error closing file: %s\n", err)
		}
	}()

	addr := fmt.Sprintf("http://%s:%s", host, port)

	// Create a new HTTPClient
	client := influxdb2.NewClientWithOptions(addr, token,
		influxdb2.DefaultOptions().SetBatchSize(20))

	logger.Printf("Connected to influxdb!")
	readPoints(config, client, logger)

	client.Close()

}
