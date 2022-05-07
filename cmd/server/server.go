package main

import (
	"github.com/Constantine-IT/devops/cmd/server/internal/handlers"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
	"log"
	"net/http"
)

func main() {
	//	конфигурация приложения через считывание флагов и переменных окружения
	cfg := newConfig()

	//	конструктор источника данных сервера, на основе входящих параметров
	datasource, syncWriter, err := storage.NewDatasource(cfg.DatabaseDSN, cfg.StoreFile, cfg.StoreInterval, cfg.RestoreOnStart)
	if err != nil {
		cfg.ErrorLog.Fatal(err)
	}

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   cfg.ErrorLog,  //	журнал ошибок приложения
		InfoLog:    cfg.InfoLog,   //	журнал информационных сообщений
		KeyToSign:  cfg.KeyToSign, //	ключ для подписи метрик по алгоритму HMAC c SHA256
		Datasource: datasource,    //	источник данных для хранения метрик
		SyncWriter: syncWriter,    //	дескриптор записи в файл-хранилище
	}

	//	при остановке сервера отложенно закроем все источники данных
	defer app.Datasource.Close()

	srv := &http.Server{
		Addr:     cfg.ServerAddress, //	адрес запуска сервера
		ErrorLog: app.ErrorLog,      //	журнал ошибок сервера
		Handler:  app.Routes(),      //	маршрутизатор сервера
	}
	//	запускаем сервер сбора метрик
	log.Fatal(srv.ListenAndServe())
}
