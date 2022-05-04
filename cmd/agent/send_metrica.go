package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"runtime"
)

//	Metrics - структура для обмена информацией о метриках между сервером и агентами мониторинга
type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string  `json:"hash,omitempty"`  // значение хеш-функции
}

func sendMetrics(m runtime.MemStats, pollCount int64, serverAddress, KeyToSign string) {

	gaugeMetrics := make(map[string]float64)

	//	заполняем массив с метриками статистикой, собранной ранее в структуру runtime.MemStats
	gaugeMetrics["Alloc"] = float64(m.Alloc)
	gaugeMetrics["BuckHashSys"] = float64(m.BuckHashSys)
	gaugeMetrics["Frees"] = float64(m.Frees)
	gaugeMetrics["GCCPUFraction"] = m.GCCPUFraction
	gaugeMetrics["GCSys"] = float64(m.GCSys)
	gaugeMetrics["HeapAlloc"] = float64(m.HeapAlloc)
	gaugeMetrics["HeapIdle"] = float64(m.HeapIdle)
	gaugeMetrics["HeapInuse"] = float64(m.HeapInuse)
	gaugeMetrics["HeapObjects"] = float64(m.HeapObjects)
	gaugeMetrics["HeapReleased"] = float64(m.HeapReleased)
	gaugeMetrics["HeapSys"] = float64(m.HeapSys)
	gaugeMetrics["LastGC"] = float64(m.LastGC)
	gaugeMetrics["Lookups"] = float64(m.Lookups)
	gaugeMetrics["MCacheInuse"] = float64(m.MCacheInuse)
	gaugeMetrics["MCacheSys"] = float64(m.MCacheSys)
	gaugeMetrics["MSpanInuse"] = float64(m.MSpanInuse)
	gaugeMetrics["MSpanSys"] = float64(m.MSpanSys)
	gaugeMetrics["Mallocs"] = float64(m.Mallocs)
	gaugeMetrics["NextGC"] = float64(m.NextGC)
	gaugeMetrics["NumForcedGC"] = float64(m.NumForcedGC)
	gaugeMetrics["NumGC"] = float64(m.NumGC)
	gaugeMetrics["OtherSys"] = float64(m.OtherSys)
	gaugeMetrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	gaugeMetrics["StackInuse"] = float64(m.StackInuse)
	gaugeMetrics["StackSys"] = float64(m.StackSys)
	gaugeMetrics["Sys"] = float64(m.Sys)
	gaugeMetrics["TotalAlloc"] = float64(m.TotalAlloc)
	gaugeMetrics["RandomValue"] = rand.Float64()

	// создаём клиента для отправки метрик на сервер
	client := resty.New()

	for name, row := range gaugeMetrics { //	пробегаем по всем метрикам типа gauge
		metrica := Metrics{ //	изготавливаем структуру для отправки данных
			ID:    name,
			MType: "gauge",
			Delta: 0,
			Value: row,
		}
		if KeyToSign != "" { //	если ключ для изготовления подписи задан, вставляем в метрику подпись SHA256
			hash256 := sha256.Sum256([]byte(fmt.Sprintf("%s:gauge:%f:key:%s", metrica.ID, metrica.Value, KeyToSign)))
			metrica.Hash = fmt.Sprintf("%X", hash256)
		}

		sendPostMetrica(metrica, client, serverAddress) //	отправляем метрику на сервер
	}

	//	пробегаем по всем метрикам типа counter

	metrica := Metrics{ //	изготавливаем структуру для отправки данных
		ID:    "PollCount",
		MType: "counter",
		Delta: pollCount,
		Value: 0,
	}
	if KeyToSign != "" { //	если ключ для изготовления подписи задан, вставляем в метрику подпись SHA256
		hash256 := sha256.Sum256([]byte(fmt.Sprintf("%s:counter:%d:key:%s", metrica.ID, metrica.Delta, KeyToSign)))
		metrica.Hash = fmt.Sprintf("%X", hash256)
	}

	sendPostMetrica(metrica, client, serverAddress) //	отправляем метрику на сервер

}

func sendPostMetrica(metrica Metrics, client *resty.Client, serverAddress string) {
	//	изготавливаем JSON
	metricsJSON, err := json.Marshal(metrica)
	if err != nil || metricsJSON == nil {
		log.Println("couldn't marshal metrica JSON")
	}

	// отправляем метрику на сервер через JSON API
	_, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(metricsJSON).
		Post("http://" + serverAddress + "/update/")
	//if err != nil {
	//	log.Println(err.Error())
	//}
}
