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
	MetricaName := chi.URLParam(r, "MetricaName")
	MetricaType := chi.URLParam(r, "MetricaType")

	if MetricaType != "gauge" && MetricaType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica types are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica types are allowed")
		return
	}

	//func (s *Storage) Get(name string) (mType string, delta int64, value float64, flg int)

	//	ищем в базее связку MetricaValue по заданным MetricaName + MetricaType
	MetricaTypeFromDB, MetricaDelta, MetricaValue, flag := app.Datasource.Get(MetricaName)

	switch flag {
	//	анализируем значение флага для выборки метрики
	case 0: //	если метрика в базе не найдена
		http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
		app.ErrorLog.Println("There is no such METRICA in our database")
		return
	case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
		if MetricaType != MetricaTypeFromDB { //	если тип метрики НЕ совпадает с хранимым в базе
			http.Error(w, "metrica type you specified is NOT the same as in database", http.StatusBadRequest)
			app.ErrorLog.Println("Metrica get error: metrica types you specified is NOT the same as in database")
			return
		} else { //	если тип метрики совпадает с хранимым в базе
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			if MetricaType == "gauge" {
				var value []byte
				value = strconv.AppendFloat(value, MetricaValue, 'f', -1, 64)
				w.Write(value) //	пишем MetricaValue в текстовом виде в тело ответа
			}
			if MetricaType == "counter" {
				var delta []byte
				delta = strconv.AppendInt(delta, MetricaDelta, 10)
				w.Write(delta) //	пишем MetricaValue в текстовом виде в тело ответа
			}
		}
	default:
		http.Error(w, "Something goes wrong", http.StatusInternalServerError)
	}
}
