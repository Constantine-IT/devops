package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

//	GetMetricaHandler - обработчик GET - возвращает значение метикрики по данным из
//	PATH = "/value/{MetricaType}/{MetricaName}"
func (app *Application) GetMetricaHandler(w http.ResponseWriter, r *http.Request) {

	//	считываем имя метрики из PATH входящего запроса
	Name := chi.URLParam(r, "MetricaName")
	Type := chi.URLParam(r, "MetricaType")

	if Type != "gauge" && Type != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
	}

	//	ищем в базее связку MetricaValue по заданным MetricaName + MetricaType
	MetricaTypeFromDB, MetricaDeltaFromDB, MetricaValueFromDB, flag := app.Datasource.Get(Name)

	switch flag {
	//	анализируем значение флага для выборки метрики
	case 0: //	если метрика в базе не найдена
		http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
		app.ErrorLog.Println("There is no such METRICA in our database")
		return
	case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
		if Type != MetricaTypeFromDB { //	если тип метрики НЕ совпадает с хранимым в базе
			http.Error(w, "metrica type you specified is NOT the same as in database", http.StatusBadRequest)
			app.ErrorLog.Println("Metrica get error: metrica types you specified is NOT the same as in database")
			return
		} else { //	если тип метрики совпадает с хранимым в базе
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			if Type == "gauge" {
				var value []byte
				value = strconv.AppendFloat(value, MetricaValueFromDB, 'f', 6, 64)
				w.Write(value) //	пишем MetricaValue в BYTE виде в тело ответа
			}
			if Type == "counter" {
				var delta []byte
				delta = strconv.AppendInt(delta, MetricaDeltaFromDB, 10)
				w.Write(delta) //	пишем MetricaValue в BYTE виде в тело ответа
			}
		}
	default:
		http.Error(w, "Something goes wrong", http.StatusInternalServerError)
	}
}
