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

	for {
		// Выполняем HTTP GET запрос
		resp, err := http.Get(url)
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(5 * time.Second)
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
			time.Sleep(5 * time.Second)
			continue
		}

		// Читаем тело ответа
		scanner := bufio.NewScanner(resp.Body)
		if scanner.Scan() {
			data := scanner.Text()
			processStats(data)
		} else {
			errorCount++
		}

		resp.Body.Close()

		// Сбрасываем счетчик ошибок при успешном запросе
		if errorCount > 0 {
			errorCount = 0
		}

		// Ждем перед следующим запросом
		time.Sleep(5 * time.Second)
	}
}

func processStats(data string) {
	// Разделяем данные по запятым
	values := strings.Split(data, ",")
	if len(values) != 7 {
		return // Неверный формат данных
	}

	// Парсим значения
	loadAvg, err1 := strconv.ParseFloat(values[0], 64)
	totalMem, err2 := strconv.ParseUint(values[1], 10, 64)
	usedMem, err3 := strconv.ParseUint(values[2], 10, 64)
	totalDisk, err4 := strconv.ParseUint(values[3], 10, 64)
	usedDisk, err5 := strconv.ParseUint(values[4], 10, 64)
	totalNet, err6 := strconv.ParseUint(values[5], 10, 64)
	usedNet, err7 := strconv.ParseUint(values[6], 10, 64)

	// Проверяем ошибки парсинга
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
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