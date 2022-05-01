package main

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
)

//type gauge float64
//type counter int64

func sendMetrica(m runtime.MemStats, pollCount uint64, serverAddress string) {

	type metricaRow struct {
		mType string
		value string
	}

	metrica := make(map[string]metricaRow)

	metrica["Alloc"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.Alloc, 10)}
	metrica["BuckHashSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.BuckHashSys, 10)}
	metrica["Frees"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.Frees, 10)}
	metrica["GCCPUFraction"] = metricaRow{mType: "gauge", value: strconv.FormatFloat(m.GCCPUFraction, 'E', -1, 64)}
	metrica["GCSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.GCSys, 10)}
	metrica["HeapAlloc"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapAlloc, 10)}
	metrica["HeapIdle"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapIdle, 10)}
	metrica["HeapInuse"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapInuse, 10)}
	metrica["HeapObjects"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapObjects, 10)}
	metrica["HeapReleased"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapReleased, 10)}
	metrica["HeapSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.HeapSys, 10)}
	metrica["LastGC"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.LastGC, 10)}
	metrica["Lookups"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.Lookups, 10)}
	metrica["MCacheInuse"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.MCacheInuse, 10)}
	metrica["MCacheSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.MCacheSys, 10)}
	metrica["MSpanInuse"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.MSpanInuse, 10)}
	metrica["MSpanSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.MSpanSys, 10)}
	metrica["Mallocs"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.Mallocs, 10)}
	metrica["NextGC"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.NextGC, 10)}
	metrica["NumForcedGC"] = metricaRow{mType: "gauge", value: strconv.FormatUint(uint64(m.NumForcedGC), 10)}
	metrica["NumGC"] = metricaRow{mType: "gauge", value: strconv.FormatUint(uint64(m.NumGC), 10)}
	metrica["OtherSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.OtherSys, 10)}
	metrica["PauseTotalNs"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.PauseTotalNs, 10)}
	metrica["StackInuse"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.StackInuse, 10)}
	metrica["StackSys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.StackSys, 10)}
	metrica["Sys"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.Sys, 10)}
	metrica["TotalAlloc"] = metricaRow{mType: "gauge", value: strconv.FormatUint(m.TotalAlloc, 10)}
	metrica["RandomValue"] = metricaRow{mType: "gauge", value: strconv.FormatUint(rand.Uint64(), 10)}
	metrica["PollCount"] = metricaRow{mType: "counter", value: strconv.FormatUint(pollCount, 10)}

	for name, row := range metrica {
		resp, err := http.Post("http://"+serverAddress+"/update/"+row.mType+"/"+name+"/"+row.value, "text/plain", nil)
		if err != nil || resp == nil {
			continue
		}
		//log.Println(name, row.value)
	}

}
