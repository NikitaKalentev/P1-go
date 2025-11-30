package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	serverURL             = "http://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold      = 30.0
	memoryUsageThreshold  = 0.8
	diskUsageThreshold    = 0.9
	networkUsageThreshold = 0.9
	retryLimit            = 3
)

func main() {
	errorCount := 0
	client := &http.Client{}

	for {
		resp, err := client.Get(serverURL)
		if err != nil {
			errorCount++
			if errorCount >= retryLimit {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			errorCount++
			if errorCount >= retryLimit {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= retryLimit {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		data := strings.TrimSpace(string(body))
		stats := strings.Split(data, ",")
		if len(stats) != 7 {
			errorCount++
			if errorCount >= retryLimit {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		processStats(stats)
		errorCount = 0
	}
}

func processStats(stats []string) {
	loadAvg, err1 := strconv.ParseFloat(stats[0], 64)
	totalMem, err2 := strconv.ParseUint(stats[1], 10, 64)
	usedMem, err3 := strconv.ParseUint(stats[2], 10, 64)
	totalDisk, err4 := strconv.ParseUint(stats[3], 10, 64)
	usedDisk, err5 := strconv.ParseUint(stats[4], 10, 64)
	totalNetwork, err6 := strconv.ParseUint(stats[5], 10, 64)
	usedNetwork, err7 := strconv.ParseUint(stats[6], 10, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return
	}

	// Проверка Load Average
	if loadAvg >= loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// Проверка использования памяти
	if totalMem > 0 {
		memUsagePercent := float64(usedMem) / float64(totalMem) * 100
		if memUsagePercent > memoryUsageThreshold*100 {
			fmt.Printf("Memory usage too high: %.0f%%\n", math.Floor(memUsagePercent))
		}
	}

	// Проверка использования диска
	if totalDisk > 0 {
		diskUsage := float64(usedDisk) / float64(totalDisk)
		if diskUsage > diskUsageThreshold {
			freeDiskMB := float64(totalDisk-usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Floor(freeDiskMB))
		}
	}

	// Проверка использования сети
	if totalNetwork > 0 {
		networkUsage := float64(usedNetwork) / float64(totalNetwork)
		if networkUsage > networkUsageThreshold {
			freeNetworkMbits := float64(totalNetwork-usedNetwork) / 1000000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Floor(freeNetworkMbits))
		}
	}
}