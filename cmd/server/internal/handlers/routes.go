package handlers

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
)

//	конфигурация приложения SERVER
type Application struct {
	ErrorLog   *log.Logger        //	журнал ошибок
	InfoLog    *log.Logger        //	журнал информационных сообщений
	Datasource storage.Datasource //	источник данных для хранения URL
}

func (app *Application) Routes() chi.Router {

	// определяем роутер chi
	r := chi.NewRouter()

	// зададим middleware для поддержки компрессии тел запросов и ответов
	r.Use(middleware.Compress(1, `text/plain`, `application/json`))
	r.Use(middleware.AllowContentEncoding(`gzip`))
	//	middleware для декомпрессии входящих пакетов
	r.Use(app.DecompressGZIP)
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	/*	Эндпоинт POST - принимает значение метрики в формате PATH = "/update/{MetricaType}/{MetricaName}/{MetricaValue}"
		Эндпоинт POST /update - принимает значение метрики в формате JSON со структурой Metrics
		Эндпоинт POST /value - принимает запрос значения метрики в формате JSON со структурой Metrics,
			с пустыми полями значения метрики, в ответ получает тот же JSON, но уже с заполненными полями
		Эндпоинт GET - возвращает значение метикрики по данным из PATH = "/value/{MetricaType}/{MetricaName}"
		Эндпоинт GET / - возвращает список всех сохраненных в базе метрик	*/
	r.Route("/", func(r chi.Router) {
		r.Post("/update/{MetricaType}/{MetricaName}/{MetricaValue}", app.PostMetricaHandler)
		r.Post("/update/", app.PostJSONMetricaHandler)
		r.Post("/update", app.PostJSONMetricaHandler)
		r.Post("/value/", app.GetJSONMetricaHandler)
		r.Post("/value", app.GetJSONMetricaHandler)
		r.Get("/value/{MetricaType}/{MetricaName}", app.GetMetricaHandler)
		r.Get("/", app.GetAllMetricsHandler)
	})

	return r
}
