package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//	GetJSONMetricaHandler - обработчик POST /value - принимает запрос значения метрики в формате JSON со структурой Metrics,
//	с пустыми полями значения метрики, в ответ получает тот же JSON, но уже с заполненными полями

func (app *Application) GetJSONMetricaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error
	jsonBody, err := io.ReadAll(r.Body) // считываем JSON из тела запроса
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body read error:" + err.Error())
		return
	}

	//	создаеём экземпляр структуры для заполнения из JSON
	metrica := Metrics{}

	//	структура Metrics используется для приема и выдачи значений метрик
	//	теги для JSON там уже описаны, так что дополнительного описания для парсинга не требуется
	//	парсим JSON и записываем результат в экземпляр структуры Metrics
	err = json.Unmarshal(jsonBody, &metrica)
	//	проверяем успешно ли парсится JSON
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body parsing error:" + err.Error())
		return
	}

	// поддерживаются только типы метрик gauge и counter
	if metrica.MType != "gauge" && metrica.MType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Try to insert metrica TYPE: ", metrica.MType, ", but only GAUGE or COUNTER are allowed")
		return
	}

	//	ищем в базее метрику с входящим именем, и выводим для неё тип и значение
	TypeFromDB, DeltaFromDB, ValueFromDB, flagIsExist := app.Datasource.Get(metrica.ID)

	switch flagIsExist { //	анализируем значение флага наличия метрики в базе
	case 0: //	если метрика в базе не найдена
		http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
		return
	case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
		if metrica.MType != TypeFromDB { //	если тип метрики НЕ совпадает с хранимым в базе
			http.Error(w, "metrica TYPE you specified is NOT the same as in database", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Something goes wrong", http.StatusBadRequest)
		return
	}

	//	если метрика в базе найдена (flagIsExist = 1), то преобразуем её структуру в JSON и вставляем в тело ответа
	//	структуру JSON дополнительно описывать не надо, так как структура Metrics уже имеет JSON теги
	metrica.Delta = DeltaFromDB
	metrica.Value = ValueFromDB
	if app.KeyToSign != "" { //	если ключ для подписи задан, вставляем в метрику подпись HMAC c SHA256
		h := hmac.New(sha256.New, []byte(app.KeyToSign)) //	создаём интерфейс подписи с хешированием
		//	формируем фразу для хеширования по разному шаблону для метрик типа counter и gauge
		if metrica.MType == "counter" {
			h.Write([]byte(fmt.Sprintf("%s:counter:%d", metrica.ID, metrica.Delta)))
		}
		if metrica.MType == "gauge" {
			h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metrica.ID, metrica.Value)))
		}
		hash256 := h.Sum(nil)                     //	вычисляем HASH-подпись
		metrica.Hash = fmt.Sprintf("%x", hash256) //	переводим всё в тип данных string и вставляем в метрику в поле HASH
	}

	type CounterMetrics struct {
		ID    string  `json:"id"`              // имя метрики
		MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta int64   `json:"delta"`           // значение метрики в случае передачи counter
		Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
		Hash  string  `json:"hash,omitempty"`  // значение хеш-подписи
	}
	counterMetrica := CounterMetrics{
		ID:    metrica.ID,
		MType: metrica.MType,
		Delta: metrica.Delta,
		Hash:  metrica.Hash,
	}
	type GaugeSendMetrics struct {
		ID    string  `json:"id"`              // имя метрики
		MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value float64 `json:"value"`           // значение метрики в случае передачи gauge
		Hash  string  `json:"hash,omitempty"`  // значение хеш-подписи
	}
	gaugeSendMetrica := GaugeSendMetrics{
		ID:    metrica.ID,
		MType: metrica.MType,
		Value: metrica.Value,
		Hash:  metrica.Hash,
	}

	//	изготавливаем JSON со структурой нашей метрики
	MetricsJSON, err := json.Marshal(counterMetrica)

	if metrica.MType == "gauge" {
		MetricsJSON, err = json.Marshal(gaugeSendMetrica)
	}

	if err != nil || MetricsJSON == nil { //	в случае ошибки преобразования, выдаем http.StatusInternalServerError
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("Marshal JSON ", err.Error())
		return
	}

	//	формируем ответ с http.StatusOK и метрикой в теле ответа в виде JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(MetricsJSON) //	пишем метрику в JSON виде в тело ответа
}
