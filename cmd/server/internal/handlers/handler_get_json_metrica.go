package handlers

import (
	"net/http"
)

//	GetJSONMetricaHandler - обработчик POST - принимает запрос значения метрики в формате JSON со структурой Metrics,
//	с пустыми полями значения метрики, в ответ получает тот же JSON, но уже с заполненными полями

func (app *Application) GetJSONMetricaHandler(w http.ResponseWriter, r *http.Request) {

}
