package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

//	GetMetricaHandler - обработчик GET - возвращает значение метикрики по данным из PATH = "/value/{Type}/{Name}"
func (app *Application) GetMetricaHandler(w http.ResponseWriter, r *http.Request) {
	//	считываем имя метрики и тип метрики из PATH входящего запроса
	Name := chi.URLParam(r, "Name")
	Type := chi.URLParam(r, "Type")

	// поддерживаются только типы метрик gauge и counter
	if Type != "gauge" && Type != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Try to insert metrica TYPE: ", Type, ", but only GAUGE or COUNTER are allowed")
		return
	}

	//	ищем в базее метрику с входящим именем - Name, и выводим по ней тип и значение
	TypeFromDB, DeltaFromDB, ValueFromDB, flagIsExist := app.Datasource.Get(Name)

	switch flagIsExist { //	анализируем значение флага наличия метрики в базе
	case 0: //	если метрика в базе не найдена
		http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
		return
	case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
		if Type != TypeFromDB { //	если тип метрики НЕ совпадает с хранимым в базе
			http.Error(w, "metrica type you specified is NOT the same as in database", http.StatusBadRequest)
			return
		} else { //	если тип метрики совпадает с хранимым в базе
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			if Type == "gauge" { //	для типа gauge преобразуем значение метрики из float64 в []byte
				var value []byte
				value = strconv.AppendFloat(value, ValueFromDB, 'f', -1, 64)
				w.Write(value) //	пишем значение метрики в тело ответа
			}
			if Type == "counter" { //	для типа counter преобразуем значение метрики из int64 в []byte
				var delta []byte
				delta = strconv.AppendInt(delta, DeltaFromDB, 10)
				w.Write(delta) //	пишем значение метрики в тело ответа
			}
		}
	default:
		http.Error(w, "Something goes wrong", http.StatusBadRequest)
		return
	}
}
