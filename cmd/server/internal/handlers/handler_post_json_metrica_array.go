package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostJSONMetricaArrayHandler - обработчик  POST /updates/ - принимает в теле запроса множество метрик в формате: []Metrics

func (app *Application) PostJSONMetricaArrayHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var err error
	jsonBody, err := io.ReadAll(r.Body) // считываем JSON из тела запроса
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body read error:" + err.Error())
		return
	}

	//	создаём массив структур Metrics, которые используются для приема и выдачи значений метрик
	//	теги для JSON там уже описаны, так что дополнительного описания для парсинга не требуется

	metricaArray := []Metrics{}

	//	парсим JSON и записываем результат в экземпляр структуры
	err = json.Unmarshal(jsonBody, &metricaArray)
	//	проверяем успешно ли парсится JSON
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body parsing error:" + err.Error())
		return
	}
	for _, metrica := range metricaArray { //	прогоняем каждую метрику из пришедшего массива на предмет валидности
		//	и вставляем в базу, если всё ОК

		//	проверяем тип метрики - допускается только gauge и counter
		if metrica.MType != "gauge" && metrica.MType != "counter" {
			app.ErrorLog.Println("Try to insert metrica TYPE: ", metrica.MType, ", but only GAUGE or COUNTER are allowed")
			continue
		}

		//	В случае, если во входящей структуре метрики явно не заданы её значения, то интерпретируем их как 0 (ноль)
		var Value float64 = 0
		var Delta int64 = 0
		if metrica.Value == nil {
			metrica.Value = &Value
		}
		if metrica.Delta == nil {
			metrica.Delta = &Delta
		}

		if app.KeyToSign != "" { //	если ключ для изготовления подписи задан, вычисляем для метрики подпись HMAC c SHA256
			h := hmac.New(sha256.New, []byte(app.KeyToSign)) //	создаём интерфейс подписи с хешированием
			//	формируем фразу для хеширования по разному шаблону для метрик типа counter и gauge
			if metrica.MType == "counter" {
				h.Write([]byte(fmt.Sprintf("%s:counter:%d", metrica.ID, *metrica.Delta)))
			}
			if metrica.MType == "gauge" {
				h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metrica.ID, *metrica.Value)))
			}
			hash256 := h.Sum(nil)                     //	вычисляем HASH для метрики
			metricaHash := fmt.Sprintf("%x", hash256) //	переводим всё в тип данных string
			if metrica.Hash != metricaHash {
				app.ErrorLog.Println("HASH signature of metrica is NOT valid for our server")
				continue
			}
		}

		//	сохраняем в базу связку Metrica (Name + Type + Delta/Value)
		//	если метрика имеет тип gauge, то передаем её в структуру хранения, как Value (type gauge float64)
		//	если метрика имеет тип counter, то передаем её в структуру хранения, как Delta (type counter int64)

		if err := app.Datasource.Insert(metrica.ID, metrica.MType, 0, *metrica.Value); err != nil {
			app.ErrorLog.Println("Metrica save ", err.Error())
			continue
		}
	}
	//	высылаем ответ - http.StatusOK
	w.WriteHeader(http.StatusOK)
}
