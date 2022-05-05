package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

//	PostMetricaHandler - обработчик POST принимает значение метрики в формате
//	PATH = "/update/{MetricaType}/{MetricaName}/{MetricaValue}"
func (app *Application) PostMetricaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//	считываем имя метрики из PATH входящего запроса
	Name := chi.URLParam(r, "Name")
	Type := chi.URLParam(r, "Type")
	Value := chi.URLParam(r, "Value")

	if Type != "gauge" && Type != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
	}
	_, errFloat := strconv.ParseFloat(Value, 64)
	_, errInt := strconv.ParseInt(Value, 10, 64)

	if errFloat != nil && errInt != nil { //	если оба парсинга дали ошибку - то входящая переменная и не gauge, и не counter
		http.Error(w, "only GAUGE or COUNTER metrica VALUES are allowed", http.StatusBadRequest)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica VALUES are allowed")
		return
	}
	//	сохраняем в базу связку Metrica (Name + Type + Delta/Value)
	var errInsert error

	if Type == "gauge" { //	если метрика имеет тип gauge, сохраняем её с delta = 0
		value, _ := strconv.ParseFloat(Value, 64)
		errInsert = app.Datasource.Insert(Name, Type, 0, value)
	}
	if Type == "counter" { //	если метрика имеет тип counter, сохраняем её с value = 0
		delta, _ := strconv.ParseInt(Value, 10, 64)
		errInsert = app.Datasource.Insert(Name, Type, delta, 0)
	}
	if errInsert != nil {
		http.Error(w, errInsert.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("Metrics save ", errInsert.Error())
		return
	}

	// Изготавливаем и возвращаем ответ cо статусом http.StatusOK
	w.WriteHeader(http.StatusOK)
}
