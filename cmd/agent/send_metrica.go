package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"runtime"
	"strconv"

	"github.com/go-resty/resty/v2"
)

//type gauge float64
//type counter int64
type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func sendMetrics(m *runtime.MemStats, pollCounter *PollCounter, serverAddress string) {

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

	// Create a resty client
	client := resty.New()

	//	высылаем на сервер все метрики типа gauge
	metrica := Metrics{MType: "gauge"}
	for name, row := range gaugeMetrics {
		metrica.ID = name
		metrica.Value = row
		sendPostMetrica(metrica, client, serverAddress)
	}

	//	высылаем на сервер все метрики типа counter
	metrica = Metrics{ID: "PollCount",
		MType: "counter",
		Delta: pollCounter.Count,
		Value: 0,
	}
	sendPostMetrica(metrica, client, serverAddress)

	//	в завершении передачи метрик, сбрасываем счетчик циклов измерения при передаче данных
	pollCounter.mutex.Lock()
	pollCounter.Count = 0
	pollCounter.mutex.Unlock()
	return
}

func sendPostMetrica(metrica Metrics, client *resty.Client, serverAddress string) {
	//	изготавливаем JSON
	metricsJSON, err := json.Marshal(metrica)
	if err != nil || metricsJSON == nil {
		log.Println("couldn't marshal metrica JSON")
	}
	//log.Println(string(metricsJSON))
	// POST JSON string
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(metricsJSON)).
		Post("http://" + serverAddress + "/update/gauge/" + metrica.ID + "/" + strconv.FormatFloat(metrica.Value, 'E', -1, 64))
	if err != nil {
		log.Println(err.Error())
	}
}
