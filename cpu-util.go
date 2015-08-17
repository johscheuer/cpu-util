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

type CPUUsage struct {
	idle  uint64
	total uint64
}

func readProcStat() (cpuUsages map[string]CPUUsage) {
	stats, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cpuUsages = make(map[string]CPUUsage)
	for _, stat := range strings.Split(string(stats), "\n") {
		fields := strings.Fields(stat)
		if strings.HasPrefix(fields[0], "cpu") {
			cpuUsage := CPUUsage{}
			for index, field := range fields {
				if index < 1 {
					continue
				}
				value, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					fmt.Println("Error at index: ", index)
					fmt.Println(err)
				}

				cpuUsage.total += value //count all ticks

				if index == 4 {
					cpuUsage.idle = value
				}
			}
			cpuUsages[fields[0]] = cpuUsage
		} else {
			return
		}

	}
	return
}

func calculateCPUUsage() {
	cpuUsages_begin := readProcStat()
	time.Sleep(time.Millisecond * 500)
	cpuUsages_end := readProcStat()

	for cpu, cpuUsage := range cpuUsages_begin {
		totalTime := float64(cpuUsages_end[cpu].total - cpuUsage.total)
		cpuUsage := 100 * (totalTime - float64(cpuUsages_end[cpu].idle-cpuUsage.idle)) / totalTime
		fmt.Printf("CPU %s usage: %.3f%%\n", cpu, cpuUsage)
	}
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
