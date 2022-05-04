package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//	PostJSONMetricaHandler - обработчик POST /update/ принимает значение метрики в формате JSON со структурой Metrics
func (app *Application) PostJSONMetricaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error
	jsonBody, err := io.ReadAll(r.Body) // считываем JSON из тела запроса
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body read error:" + err.Error())
		return
	}

	//	структура storage.Metrics используется для приема и выдачи значений метрик
	//	теги для JSON там уже описаны, так что дополнительного описания для парсинга не требуется

	metrica := Metrics{}

	//	парсим JSON и записываем результат в экземпляр структуры
	err = json.Unmarshal(jsonBody, &metrica)
	//	проверяем успешно ли парсится JSON
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body parsing error:" + err.Error())
		return
	}

	//	проверяем тип метрики - допускается только gauge и counter
	if metrica.MType != "gauge" && metrica.MType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
	}

	if app.KeyToSign != "" { //	если ключ для подписи метрик задан на сервере, проверяем подпись входящей метрики
		var hash256 [32]byte
		if metrica.MType == "counter" {
			hash256 = sha256.Sum256([]byte(fmt.Sprintf("%s:counter:%d:key:%s", metrica.ID, metrica.Delta, app.KeyToSign)))
		}
		if metrica.MType == "gauge" {
			hash256 = sha256.Sum256([]byte(fmt.Sprintf("%s:gauge:%f:key:%s", metrica.ID, metrica.Value, app.KeyToSign)))
		}
		metricaHash := fmt.Sprintf("%X", hash256)
		if metrica.Hash != metricaHash {
			http.Error(w, "HASH signature of metrica is NOT valid for uor server", http.StatusBadRequest)
			app.ErrorLog.Println("HASH signature of metrica is NOT valid for uor server")
			return
		}
	}

	//	сохраняем в базу связку MetricaName + MetricaType + MetricaValue
	//	если метрика имеет тип gauge, то передаем её в структуру хранения, как Value - type gauge float64
	//	если метрика имеет тип counter, то передаем её в структуру хранения, как Delta - type counter int64

	var errType error

	if metrica.MType == "gauge" {
		errType = app.Datasource.Insert(metrica.ID, metrica.MType, 0, metrica.Value)
	}
	if metrica.MType == "counter" {
		errType = app.Datasource.Insert(metrica.ID, metrica.MType, metrica.Delta, 0)
	}
	if errType != nil {
		http.Error(w, errType.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println("URL save error:" + errType.Error())
		return
	}

	//	высылаем ответ - http.StatusOK
	w.WriteHeader(http.StatusOK)
}
