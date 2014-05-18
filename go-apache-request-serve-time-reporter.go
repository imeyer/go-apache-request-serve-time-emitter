package main

import (
	"bufio"
	"flag"
	"github.com/influxdb/influxdb-go"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var interval = flag.String("interval", "15s", "Interval to process data and send to the configured writer")

func median(numbers []float64) int64 {
	if len(numbers) == 0 {
		return 0
	}
	start_index := 0
	end_index := len(numbers) - 1
	median := int64((start_index + end_index) / 2)
	return int64(numbers[median])
}

func main() {
	flag.Parse()
	interval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Panic("ERROR: ", err)
	}
	ticker := time.NewTicker(interval)
	reader := bufio.NewReader(os.Stdin)
	quit := make(chan struct{})
	numbers := make(chan float64)

	go func() {
		for {
			value, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Can not read string.\n")
			}
			value = strings.TrimSuffix(value, "\n")
			num, err := strconv.ParseFloat(value, 64)

			if err != nil {
				log.Printf("Can not add number \"%d\" to array.\n")
			}
			numbers <- num
		}
	}()

	data := make([]float64, 0)
	for {
		select {
		case <-ticker.C:
			log.Printf("Average: %v\n", median(data))
			data = data[:0]
		case number := <-numbers:
			data = append(data, number)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
