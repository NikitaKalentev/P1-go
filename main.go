package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL             = "http://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold      = 30.0
	memoryUsageThreshold  = 0.8
	diskUsageThreshold    = 0.9
	networkUsageThreshold = 0.9
	retryLimit            = 3
	pollInterval          = 5 * time.Second
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil {
			errorCount++
			handleError(&errorCount)
			time.Sleep(pollInterval)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			resp.Body.Close()
			handleError(&errorCount)
			time.Sleep(pollInterval)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		if scanner.Scan() {
			data := scanner.Text()
			stats := strings.Split(strings.TrimSpace(data), ",")
			if len(stats) == 7 {
				processStats(stats)
				errorCount = 0
			} else {
				errorCount++
				handleError(&errorCount)
			}
		} else {
			errorCount++
			handleError(&errorCount)
		}

		resp.Body.Close()
		time.Sleep(pollInterval)
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
	if loadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// Проверка использования памяти (проценты)
	if totalMem > 0 {
		memUsagePercent := float64(usedMem) / float64(totalMem) * 100
		if memUsagePercent > memoryUsageThreshold*100 {
			// Округляем до целого процента
			fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(memUsagePercent))
		}
	}

	// Проверка использования диска
	if totalDisk > 0 {
		diskUsage := float64(usedDisk) / float64(totalDisk)
		if diskUsage > diskUsageThreshold {
			// Свободное место в мегабайтах (1 MB = 1024 * 1024 bytes)
			freeDiskMB := float64(totalDisk-usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Round(freeDiskMB))
		}
	}

	// Проверка использования сети
	if totalNetwork > 0 {
		networkUsage := float64(usedNetwork) / float64(totalNetwork)
		if networkUsage > networkUsageThreshold {
			// Свободная полоса в мегабитах в секунду (1 Mbit = 1,000,000 bits)
			// usedNetwork и totalNetwork в байтах в секунду
			freeNetworkBytes := totalNetwork - usedNetwork
			freeNetworkMbits := float64(freeNetworkBytes) * 8 / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Round(freeNetworkMbits))
		}
	}
}

func handleError(errorCount *int) {
	if *errorCount >= retryLimit {
		fmt.Println("Unable to fetch server statistic")
	}
}