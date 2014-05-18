package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/influxdb/influxdb-go"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var interval = flag.String("interval", "15s", "Interval to process data before sending it to InfluxDB")
var metric_prefix = flag.String("metric-prefix", "", "Set your own prefix to 'apache'")
var influxdb_host = flag.String("influxdb-host", "localhost", "Hostname of InfluxDB server")
var influxdb_port = flag.Int("influxdb-port", 8086, "Port of InfluxDB server")
var influxdb_username = flag.String("influxdb-username", "root", "Username to connect to InfluxDB server")
var influxdb_password = flag.String("influxdb-password", "", "noop flag. Set INFLUXDB_PASSWORD environment variable instead.")
var influxdb_database = flag.String("influxdb-database", "database", "Database on the InfluxDB server")

func ReverseHostname(hostname string) string {

	chunks := strings.Split(hostname, ".")

	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	reverse_hostname := strings.Join(chunks[2:], ".")

	return reverse_hostname
}

func median(numbers []float64) int64 {
	if len(numbers) == 0 {
		return 0
	}
	start_index := 0
	end_index := len(numbers) - 1
	median := int64((start_index + end_index) / 2)
	return int64(numbers[median])
}

func MetricPrefix(metric_prefix string) string {
	if strings.TrimSpace(metric_prefix) != "" {
		return fmt.Sprintf("%s.apache", strings.TrimSpace(metric_prefix))
	}
	return "apache"
}

func main() {
	var memory_stats runtime.MemStats

	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		log.Panic("os.Hostname() failed: %s", err)
	}

	interval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Panic("Error parsing interval duration: ", err)
	}

	ticker := time.NewTicker(interval)
	reader := bufio.NewReader(os.Stdin)
	quit := make(chan struct{})
	numbers := make(chan float64)
	data := make([]float64, 0)

	influxdb_host_port := fmt.Sprintf("%s:%d", *influxdb_host, *influxdb_port)

	influxdb_client, err := influxdb.NewClient(&influxdb.ClientConfig{
		Host:     influxdb_host_port,
		Username: *influxdb_username,
		Password: os.Getenv("INFLUXDB_PASSWORD"),
		Database: *influxdb_database,
	})

	influxdb_hostname := fmt.Sprintf("%s.%s", MetricPrefix(*metric_prefix), ReverseHostname(hostname))

	go func() {
		for {
			value, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Nothing to be read.")
			}
			value = strings.TrimSpace(value)

			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Printf("Can not convert \"%v\" to a float: %v.\n", num, err)
			}
			numbers <- num
		}
	}()

	for {
		select {
		case <-ticker.C:
			median := median(data)
			requests := len(data)
			runtime.ReadMemStats(&memory_stats)
			series := &influxdb.Series{
				Name:    influxdb_hostname,
				Columns: []string{"request_time_median", "requests", "go_HeapAlloc", "go_HeapIdle", "go_HeapReleased", "go_Alloc"},
				Points: [][]interface{}{
					[]interface{}{median, requests, memory_stats.HeapAlloc, memory_stats.HeapIdle, memory_stats.HeapReleased, memory_stats.Alloc},
				},
			}

			if err := influxdb_client.WriteSeries([]*influxdb.Series{series}); err != nil {
				log.Panicf("Could not write %v to %v: %v", series, influxdb_hostname, err)
			}
			log.Printf("Sent %s{%f, %f, %f, %f, %f, %f} to %s", influxdb_hostname, median, requests, memory_stats.HeapSys, memory_stats.HeapAlloc, memory_stats.HeapIdle, memory_stats.HeapReleased, memory_stats.Alloc, influxdb_host_port)

			data = data[:0]
		case number := <-numbers:
			data = append(data, number)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
