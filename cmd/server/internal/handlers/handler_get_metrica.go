package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

//	GetMetricaHandler - обработчик GET - возвращает значение метикрики по данным из
//	PATH = "/value/{MetricaType}/{MetricaName}"
func (app *Application) GetMetricaHandler(w http.ResponseWriter, r *http.Request) {

	//	считываем имя метрики из PATH входящего запроса
	MetricaName := chi.URLParam(r, "MetricaName")
	MetricaType := chi.URLParam(r, "MetricaType")

	if MetricaType != "gauge" && MetricaType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica types are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica types are allowed")
		return
	}

	//	ищем в базее связку MetricaValue по заданным MetricaName + MetricaType
	MetricaTypeFromDB, MetricaValue, flag := app.Datasource.Get(MetricaName)

	switch flag {
	//	анализируем значение флага для выборки метрики
	case 0: //	если метрика в базе не найдена
		http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
		app.ErrorLog.Println("There is no such METRICA in our database")
		return
	case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
		if MetricaType == MetricaTypeFromDB { //	если тип метрики совпал
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(MetricaValue)) //	пишем MetricaValue в текстовом виде в тело ответа
		} else { //	если тип для метрики не совпал с хранимым в базе
			http.Error(w, "metrica type you specified is NOT the same as in database", http.StatusBadRequest)
			app.ErrorLog.Println("Metrica get error: metrica types you specified is NOT the same as in database")
			return
		}
	default:
		http.Error(w, "Something goes wrong", http.StatusInternalServerError)
	}
}
