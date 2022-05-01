package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

//	PostMetricaHandler - обработчик POST принимает значение метрики в формате
//	PATH = "/update/{MetricaType}/{MetricaName}/{MetricaValue}"
func (app *Application) PostMetricaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//	считываем имя метрики из PATH входящего запроса
	MetricaName := chi.URLParam(r, "MetricaName")
	MetricaType := chi.URLParam(r, "MetricaType")
	MetricaValue := chi.URLParam(r, "MetricaValue")

	if MetricaType != "gauge" && MetricaType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica types are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica types are allowed")
		return
	}
	if _, err := strconv.ParseFloat(MetricaValue, 64); err != nil {
		if _, err := strconv.ParseUint(MetricaValue, 10, 64); err != nil {
			http.Error(w, "only GAUGE or COUNTER metrica values are allowed", http.StatusBadRequest)
			app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica values are allowed")
			return
		}
	}
	//	сохраняем в базу связку MetricaName + MetricaType + MetricaValue
	err := app.Datasource.Insert(MetricaName, MetricaType, MetricaValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println("URL save error:" + err.Error())
		return
	}

	// Изготавливаем и возвращаем ответ, вставляя короткий URL в тело ответа в виде текста
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
