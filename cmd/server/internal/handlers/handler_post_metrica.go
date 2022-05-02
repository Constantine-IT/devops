package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
	}
	_, errFloat := strconv.ParseFloat(MetricaValue, 64)
	_, errInt := strconv.ParseInt(MetricaValue, 10, 64)
	if errFloat != nil || errInt != nil {
		http.Error(w, "only GAUGE or COUNTER metrica VALUES are allowed", http.StatusBadRequest)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica VALUES are allowed")
		return
	}
	//	сохраняем в базу связку MetricaName + MetricaType + MetricaValue
	//	если метрика имеет тип gauge, то передаем её в структуру хранения, как Value - type gauge float64
	//	если метрика имеет тип counter, то передаем её в структуру хранения, как Delta - type counter int64
	var err error
	log.Println("OLD SCHOOL method", MetricaName, MetricaType, MetricaValue)

	if MetricaType == "gauge" {
		value, _ := strconv.ParseFloat(MetricaValue, 64)
		err = app.Datasource.Insert(MetricaName, MetricaType, 0, value)
	}
	if MetricaType == "counter" {
		delta, _ := strconv.ParseInt(MetricaValue, 10, 64)
		err = app.Datasource.Insert(MetricaName, MetricaType, delta, 0)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println("URL save error:" + err.Error())
		return
	}
	log.Println("OLD SCHOOL method insert SUCCESSFUL")
	// Изготавливаем и возвращаем ответ c http.StatusOK
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
