package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func readProcStat() (idle, total uint64) {
	stats, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, stat := range strings.Split(string(stats), "\n") {
		fields := strings.Fields(stat)
		if fields[0] == "cpu" {
			for index, field := range fields {
				if index < 1 {
					continue
				}
				value, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					fmt.Println("Error at index: ", index)
					fmt.Println(err)
				}

				total += value //count all ticks

				if index == 4 {
					idle = value
				}
			}
		}
		return
	}
	return
}

func calculateCPUUsage() {
	idle_begin, total_begin := readProcStat()
	time.Sleep(time.Millisecond * 500)
	idle_end, total_end := readProcStat()

	totalTime := float64(total_end - total_begin)
	cpuUsage := 100 * (totalTime - float64(idle_end-idle_begin)) / totalTime

	fmt.Printf("CPU usage: %.3f%%\n", cpuUsage)
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			fmt.Printf("Exit cpu-util with signal %v", sig)
			os.Exit(1)
		}
	}()

	ticker := time.NewTicker(time.Second * 2) //tick every 2 seconds
	quit := make(chan struct{})
	calculateCPUUsage()

	go func() {
		for {
			select {
			case <-ticker.C:
				calculateCPUUsage()
			case <-quit:
				ticker.Stop()
				os.Exit(0)
			}
		}
	}()
	for {
		time.Sleep(time.Minute * 1500)
	}
}
