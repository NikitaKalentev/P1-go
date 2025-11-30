
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"
	errorCount := 0
	maxErrors := 3

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Выполняем HTTP GET запрос
		resp, err := http.Get(url)
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		// Проверяем статус ответа
		if resp.StatusCode != http.StatusOK {
			errorCount++
			resp.Body.Close()
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		// Читаем тело ответа
		scanner := bufio.NewScanner(resp.Body)
		if scanner.Scan() {
			data := scanner.Text()
			processStats(data)
			// Сбрасываем счетчик ошибок при успешном запросе
			errorCount = 0
		} else {
			errorCount++
		}

		resp.Body.Close()

		if errorCount >= maxErrors {
			fmt.Println("Unable to fetch server statistic")
			return
		}
	}
}

func processStats(data string) {
	// Разделяем данные по запятым
	values := strings.Split(strings.TrimSpace(data), ",")
	if len(values) != 7 {
		return // Неверный формат данных
	}

	// Парсим значения с проверкой ошибок
	var loadAvg float64
	var totalMem, usedMem, totalDisk, usedDisk, totalNet, usedNet uint64
	var parseErr error

	if loadAvg, parseErr = strconv.ParseFloat(values[0], 64); parseErr != nil {
		return
	}
	if totalMem, parseErr = strconv.ParseUint(values[1], 10, 64); parseErr != nil {
		return
	}
	if usedMem, parseErr = strconv.ParseUint(values[2], 10, 64); parseErr != nil {
		return
	}
	if totalDisk, parseErr = strconv.ParseUint(values[3], 10, 64); parseErr != nil {
		return
	}
	if usedDisk, parseErr = strconv.ParseUint(values[4], 10, 64); parseErr != nil {
		return
	}
	if totalNet, parseErr = strconv.ParseUint(values[5], 10, 64); parseErr != nil {
		return
	}
	if usedNet, parseErr = strconv.ParseUint(values[6], 10, 64); parseErr != nil {
		return
	}

	// Проверяем Load Average
	if loadAvg > 30 {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}

	// Проверяем использование памяти
	if totalMem > 0 {
		memUsagePercent := float64(usedMem) / float64(totalMem) * 100
		if memUsagePercent > 80 {
			fmt.Printf("Memory usage too high: %.2f%%\n", memUsagePercent)
		}
	}

	// Проверяем свободное место на диске
	if totalDisk > 0 {
		diskUsagePercent := float64(usedDisk) / float64(totalDisk) * 100
		if diskUsagePercent > 90 {
			freeSpaceMB := float64(totalDisk-usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeSpaceMB)
		}
	}

	// Проверяем использование сети
	if totalNet > 0 {
		netUsagePercent := float64(usedNet) / float64(totalNet) * 100
		if netUsagePercent > 90 {
			availableBandwidthMbit := float64(totalNet-usedNet) * 8 / (1000 * 1000)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", availableBandwidthMbit)
		}
	}
}