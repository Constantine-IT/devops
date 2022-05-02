package handlers

import (
	"encoding/json"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
	"io"
	"net/http"
)

//	GetJSONMetricaHandler - обработчик POST /value - принимает запрос значения метрики в формате JSON со структурой Metrics,
//	с пустыми полями значения метрики, в ответ получает тот же JSON, но уже с заполненными полями

func (app *Application) GetJSONMetricaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var err error
	jsonBody, err := io.ReadAll(r.Body) // считываем JSON из тела запроса
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body read error:" + err.Error())
		return
	}

	//	структура storage.Metrics используется для приема и выдачи значений метрик
	//	теги для JSON там уже описаны, так что дополнительного описания для парсинга не требуется

	//	создаеём экземпляр структуры для заполнения из JSON
	metrica := storage.Metrics{}

	//	парсим JSON и записываем результат в экземпляр структуры
	err = json.Unmarshal(jsonBody, &metrica)
	//	проверяем успешно ли парсится JSON
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		app.ErrorLog.Println("JSON body parsing error:" + err.Error())
		return
	}
	//log.Println("incoming ==>", metrica)
	if metrica.MType != "gauge" && metrica.MType != "counter" {
		http.Error(w, "only GAUGE or COUNTER metrica TYPES are allowed", http.StatusNotImplemented)
		app.ErrorLog.Println("Metrica save error: only GAUGE or COUNTER metrica TYPES are allowed")
		return
	}

	_, metrica.Delta, metrica.Value, _ = app.Datasource.Get(metrica.ID)
	//log.Println("outgoing ==>", metrica)
	//metrica.Delta = MetricaDeltaFromDB
	//metrica.Value = MetricaValueFromDB

	/*
		switch flagIsExist {
		//	анализируем значение флага для выборки метрики
		case 0: //	если метрика в базе не найдена
			http.Error(w, "There is no such METRICA in our database", http.StatusNotFound)
			app.ErrorLog.Println("There is no such METRICA in our database")
			return
		case 1: //	если метрика в базе найдена, то проверяем, того ли она типа, что указывалось при её сохранении
			if metrica.MType != MetricaTypeFromDB { //	если тип метрики НЕ совпадает с хранимым в базе
				http.Error(w, "metrica TYPE you specified is NOT the same as in database", http.StatusBadRequest)
				app.ErrorLog.Println("metrica TYPE you specified is NOT the same as in database")
				return
			}
		}
	*/

	//	если метрика в базе найдена, то преобразуем её структуру в JSON и вставляем в тело ответа
	//	структуру JSON дополнительно описывать не надо, так как возвращаемая функцией Get структура Metrics уже имеет JSON теги
	metricsJSON, err := json.Marshal(metrica) //	изготавливаем JSON
	if err != nil || metricsJSON == nil {     //	в случае ошибки преобразования, выдаем http.StatusInternalServerError
		http.Error(w, err.Error(), http.StatusInternalServerError)
		app.ErrorLog.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(metricsJSON) //	пишем MetricaValue в JSON виде в тело ответа

}
