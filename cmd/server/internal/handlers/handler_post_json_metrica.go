package handlers

import (
	"encoding/json"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
	"io"
	"net/http"
)

//	PostJSONMetricaHandler - обработчик POST принимает значение метрики в формате JSON со структурой Metrics
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

	metrica := storage.Metrics{}

	//	парсим JSON и записываем результат в экземпляр структуры
	err = json.Unmarshal(jsonBody, &metrica)
	//	проверяем успешно ли парсится JSON
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body parsing error:" + err.Error())
		return
	}

	if metrica.MType != "gauge" && metrica.MType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
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

	// Изготавливаем и возвращаем ответ c http.StatusOK
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
