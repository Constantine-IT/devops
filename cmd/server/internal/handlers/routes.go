package handlers

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
)

type Application struct {
	ErrorLog   *log.Logger        //	журнал ошибок
	InfoLog    *log.Logger        //	журнал информационных сообщений
	BaseURL    string             //	базоовый адрес сервера
	Datasource storage.Datasource //	источник данных для хранения URL
}

func (app *Application) Routes() chi.Router {

	// определяем роутер chi
	r := chi.NewRouter()

	// зададим middleware для поддержки компрессии тел запросов и ответов
	r.Use(middleware.Compress(1, `text/plain`, `application/json`))
	r.Use(middleware.AllowContentEncoding(`gzip`))
	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//	Эндпоинт POST / принимает значение метрики в формате PATH = /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	r.Route("/", func(r chi.Router) {
		r.Post("/update/{MetricaType}/{MetricaName}/{MetricaValue}", app.PostMetricaHandler)

		//		r.Post("/", app.DefaultHandler)
	})

	return r
}
