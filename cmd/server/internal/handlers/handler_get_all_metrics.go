package handlers

import (
	"encoding/json"
	"net/http"
)

//	GetAllMetricsHandler - обработчик GET / - возвращает список всех сохраненных в базе метрик
func (app *Application) GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	//	ищем в базее все сохранённые связки MetricaValue + MetricaName
	metricaValues, isFound := app.Datasource.GetAll()

	switch isFound {
	//	анализируем значение флага для выдачи всех сохраненных метрик
	case false: //	если метрики в базе не найдены
		http.Error(w, "There is no METRICA in our database", http.StatusNotFound)
		app.ErrorLog.Println("There is no METRICA in our database")
		return
	case true: //	если метрика в базе найдена, то преобразуем её в JSON и вставляем в тело ответа
		//	структуру JSON дополнительно описывать не надо, так как возвращаемый функцией GetAll список уже имеет JSON теги
		metricaValuesJSON, err := json.Marshal(metricaValues) //	изготавливаем JSON
		//log.Println(string(metricaValuesJSON))
		if err != nil || metricaValuesJSON == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			app.ErrorLog.Println(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(metricaValuesJSON) //	пишем MetricaValue в JSON виде в тело ответа
	}
}
