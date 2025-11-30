package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"
)

const (
    serverURL = "http://srv.msk01.gigacorp.local/_stats"
    loadAvgThreshold = 30.0
    memoryUsageThreshold = 0.8
    diskUsageThreshold = 0.9
    networkUsageThreshold = 0.9
    retryLimit = 3
    pollInterval = 60 * time.Second
)

func main() {
    errorCount := 0
    
    for {
        resp, err := http.Get(serverURL)
        if err != nil {
            errorCount++
            handleError(&errorCount)
            continue
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            errorCount++
            handleError(&errorCount)
            continue
        }
        
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            errorCount++
            handleError(&errorCount)
            continue
        }
        
        stats := strings.Split(string(body), ",")
        if len(stats) != 6 {
            errorCount++
            handleError(&errorCount)
            continue
        }
        
        errorCount = 0
        
        loadAvg, _ := strconv.ParseFloat(stats[0], 64)
        totalMem, _ := strconv.ParseFloat(stats[1], 64)
        usedMem, _ := strconv.ParseFloat(stats[2], 64)
        totalDisk, _ := strconv.ParseFloat(stats[3], 64)
        usedDisk, _ := strconv.ParseFloat(stats[4], 64)
        totalNetwork, _ := strconv.ParseFloat(stats[5], 64)
        usedNetwork, _ := strconv.ParseFloat(stats[5], 64)
        
        // Проверка Load Average
        if loadAvg > loadAvgThreshold {
            fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
        }
        
        // Проверка использования памяти
        memUsage := usedMem / totalMem
        if memUsage > memoryUsageThreshold {
            fmt.Printf("Memory usage too high: %.2f%%\n", memUsage * 100)
        }
        
        // Проверка использования диска
        diskUsage := usedDisk / totalDisk
        if diskUsage > diskUsageThreshold {
            freeDisk := (totalDisk - usedDisk) / (1024 * 1024)
            fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeDisk)
        }
        
        // Проверка использования сети
        networkUsage := usedNetwork / totalNetwork
        if networkUsage > networkUsageThreshold {
            freeNetwork := (totalNetwork - usedNetwork) * 8 / 1000000
            fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeNetwork)
        }
        
        time.Sleep(pollInterval)
    }
}

func handleError(errorCount *int) {
    *errorCount++
    if *errorCount >= retryLimit {
        fmt.Println("Unable to fetch server statistic")
        *errorCount = 0
    }
    time.Sleep(pollInterval)
}
