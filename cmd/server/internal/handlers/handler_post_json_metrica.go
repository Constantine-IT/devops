package handlers

import (
	"crypto/hmac"
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
		app.ErrorLog.Println("Try to insert metrica TYPE: ", metrica.MType, ", but only GAUGE or COUNTER are allowed")
		return
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
			http.Error(w, "HASH signature of metrica is NOT valid for our server", http.StatusBadRequest)
			app.ErrorLog.Println("HASH signature of metrica is NOT valid for our server")
			return
		}
	}

	//	сохраняем в базу связку Metrica (Name + Type + Delta/Value)
	//	если метрика имеет тип gauge, то передаем её в структуру хранения, как Value (type gauge float64)
	//	если метрика имеет тип counter, то передаем её в структуру хранения, как Delta (type counter int64)

	if err := app.Datasource.Insert(metrica.ID, metrica.MType, *metrica.Delta, *metrica.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println("Metrics save ", err.Error())
		return
	}

	//	высылаем ответ - http.StatusOK
	w.WriteHeader(http.StatusOK)
}
