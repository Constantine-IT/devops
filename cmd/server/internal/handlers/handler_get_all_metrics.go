package handlers

import (
	"encoding/json"
	"net/http"
)

//	MetricaValue - структура для выдачи списка всех сохранённых метрик по запросу
//	используется методах Storage.GetAll и GetAllMetricsHandler
type MetricaValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//	GetAllMetricsHandler - обработчик GET / - возвращает список всех сохраненных в базе метрик
func (app *Application) GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	//	ищем в базее все сохранённые связки MetricaValue + MetricaName
	metrics := app.Datasource.GetAll()

	if len(metrics) == 0 {
		//	если метрики в базе не найдены
		http.Error(w, "There is no METRICA in our database", http.StatusNotFound)
		app.ErrorLog.Println("There is no METRICA in our database")
		return
	}

	//	если метрики в базе найдены, то преобразуем массив с ними в JSON и вставляем в тело ответа
	//	структуру JSON дополнительно описывать не надо, так как возвращаемый функцией GetAll список уже имеет JSON теги
	metricsJSON, err := json.Marshal(metrics) //	изготавливаем JSON
	if err != nil || metricsJSON == nil {     //	в случае ошибки преобразования, выдаем http.StatusInternalServerError
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println(err.Error())
		return
	}
	//	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(metricsJSON) //	пишем MetricaValue в JSON виде в тело ответа
}
