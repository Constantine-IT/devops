package handlers

import (
	"fmt"
	"net/http"
)

//	GetAllMetricsHandler - обработчик GET / - возвращает список всех сохраненных в базе метрик
func (app *Application) GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	//	ищем в базе все сохранённые связки (NAME + Type + VALUE/DELTA)
	metrics := app.Datasource.GetAll() //	они возвращаются в виде слайса структур хранения - Storage

	//	создадим текстовый массив, содержащий все названия метрик и их значения
	body := make([]byte, 0)
	for i := range metrics {
		if metrics[i].MType == "gauge" { //	для метрик типа GAUGE выводим значение поля Value
			body = append(body, []byte(fmt.Sprintf("Metrica: %s = %v\n", metrics[i].ID, metrics[i].Value))...)
		}
		if metrics[i].MType == "counter" { //	для метрик типа COUNTER выводим значение поля Delta
			body = append(body, []byte(fmt.Sprintf("Metrica: %s = %v\n", metrics[i].ID, metrics[i].Delta))...)
		}
	}

	//	изготавливаем ответ - в виде текстового файла со списком метрик
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK) //	статус 200
	w.Write(body)                //	пишем метрики в тело ответа
}
