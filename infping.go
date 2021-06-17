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

	client "github.com/influxdata/influxdb1-client/v2"
	toml "github.com/pelletier/go-toml"
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

func readPoints(config *toml.Tree, con client.Client, logger *log.Logger) {
	nodes := config.Get("hosts.hosts").([]interface{})
	args := []string{"-B 1", "-D", "-r0", "-O 0", "-Q 10", "-p 1000", "-l"}
	list := []string{}
	for _, u := range nodes {
		ip, _ := u.(string)
		args = append(args, ip)
		list = append(list, ip)

	}

	logger.Printf("Going to ping the following ips: %v", list)
	cmd := exec.Command("/usr/sbin/fping", args...)

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
			writePoints(config, logger, con, ip, sent, recv, lossp, min, avg, max)
		}

	}
	std := bufio.NewReader(stdout)
	line, err := std.ReadString('\n')
	perr(err)
	logger.Printf("stdout:%s", line)
}

func writePoints(config *toml.Tree, logger *log.Logger, con client.Client, ip string, sent string, recv string, lossp string, min string, avg string, max string) {
	db := config.Get("influxdb.db").(string)
	ms := config.Get("influxdb.measurement").(string)
	ps := config.Get("influxdb.precision").(string)
	rp := config.Get("influxdb.retentionpolicy").(string)

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

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        db,
		Precision:       ps,
		RetentionPolicy: rp,
	})
	herr(err)

	// Create a point and add to batch
	tags := map[string]string{
		"addr": ip,
	}

	pt, err := client.NewPoint(ms, tags, fields, time.Now())
	herr(err)
	bp.AddPoint(pt)

	// Write the batch
	err = con.Write(bp)
	herr(err)
}

func main() {
	config, err := toml.LoadFile(path)
	herr(err)

	host := config.Get("influxdb.host").(string)
	port := config.Get("influxdb.port").(string)
	username := config.Get("influxdb.user").(string)
	password := config.Get("influxdb.pass").(string)

	logfile := config.Get("logs.logfile").(string)

	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)

	addr := fmt.Sprintf("http://%s:%s", host, port)

	// Create a new HTTPClient
	con, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	herr(err)

	dur, ver, err := con.Ping(1)
	herr(err)

	logger.Printf("Connected to influxdb! (dur:%v, ver:%s)", dur, ver)
	readPoints(config, con, logger)

}
