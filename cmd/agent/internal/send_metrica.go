package internal

import (
	"fmt"
	"log"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"math/rand"
	"runtime"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/mem"
)

//	Metrics - структура для обмена информацией о метриках между сервером и агентами мониторинга
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

//	обратите внимание, что для полей Delta и Value используется указатель на примитив, а не сам примитив
//	это сделано специально, так как в этом случае при сериализации в JSON, если значение поля не задано -
//	оно в JSON вообще не попадёт, а если задано, пусть даже равным 0 (нулю) - попадёт в явном виде

func SendMetrics(m runtime.MemStats, g mem.VirtualMemoryStat, pollCount int64, serverAddress, KeyToSign string) {

	// создаём срез для хранения метрик
	gaugeMetrics := make(map[string]float64)
	MetricaArray := make([]Metrics, 0)

	//	заполняем срез метриками типа gauge из статистики,
	//	собранной ранее в структуры runtime.MemStats и mem.VirtualMemoryStat

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
	gaugeMetrics["TotalMemory"] = float64(g.Total)
	gaugeMetrics["FreeMemory"] = float64(g.Free)
	gaugeMetrics["CPUutilization1"] = g.UsedPercent
	gaugeMetrics["RandomValue"] = rand.Float64()

	for name := range gaugeMetrics { //	пробегаем по всем метрикам типа gauge

		var value = gaugeMetrics[name] //	компиллятор не может напрямую найти адрес ячейки map - gaugeMetrics
		//	поэтому вводим дополнительную переменную для хранения нужного значения
		//	причём при каждом прогоне цикла - это будет новая переменная с новым адресом
		//	иначе значение переменной перезаписывалось бы при каждом прогоне цикла

		//	изготавливаем структуру для отправки метрик типа gauge
		metricaGauge := Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		if KeyToSign != "" { //	если ключ для подписи задан
			h := hmac.New(sha256.New, []byte(KeyToSign)) //	создаём интерфейс подписи с хешированием
			//	формируем фразу для хеширования метрики по шаблону типа gauge
			h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricaGauge.ID, *metricaGauge.Value)))
			hash256 := h.Sum(nil) //	вычисляем HASH сумму в виде []byte
			//	переводим её в тип данных string и вставляем в метрику подпись HMAC c SHA256
			metricaGauge.Hash = fmt.Sprintf("%x", hash256)
		}

		//	добавляем сформированную метрику в массив для отправки на сервер
		MetricaArray = append(MetricaArray, metricaGauge)
	}

	//	теперь пройдёмся по всем метрикам типа counter

	//	изготавливаем структуру для отправки метрик типа counter,
	//	такая метрика у нас одна, так что задаем её значения напрямую
	metricaCounter := Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &pollCount,
	}

	if KeyToSign != "" { //	если ключ для изготовления подписи задан
		h := hmac.New(sha256.New, []byte(KeyToSign)) //	создаём интерфейс подписи с хешированием
		//	формируем фразу для хеширования метрики по шаблону типа counter
		h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricaCounter.ID, *metricaCounter.Delta)))
		hash256 := h.Sum(nil) //	вычисляем HASH сумму в виде []byte
		//	переводим её в тип данных string и вставляем в метрику подпись HMAC c SHA256
		metricaCounter.Hash = fmt.Sprintf("%x", hash256)
	}

	//	добавляем сформированную метрику в массив для отправки на сервер
	MetricaArray = append(MetricaArray, metricaCounter)

	//	если массив с метриками содержит данные, то отправляем его на сервер с указанным адресом
	if len(MetricaArray) > 0 {
		sendPostMetrica(MetricaArray, serverAddress)
	}

}

//	sendPostMetrica - функция отправки массива метрик на указанный серверный адрес
func sendPostMetrica(MetricaArray []Metrics, serverAddress string) {
	// создаём HTTP-клиента для отправки метрик на сервер
	client := resty.New()

	//	изготавливаем JSON
	metricsJSON, err := json.Marshal(MetricaArray)
	if err != nil || metricsJSON == nil {
		log.Println("couldn't marshal metrica JSON")
	}

	// отправляем метрику на сервер через JSON API
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(metricsJSON).
		Post("http://" + serverAddress + "/updates/")
	if err != nil {
		log.Println(err.Error())
	}
}
