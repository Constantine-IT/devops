package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Constantine-IT/devops/cmd/server/internal/handlers"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
)

func main() {
	//	конфигурация приложения через считывание флагов и переменных окружения
	cfg := newConfig()

	//	конструктор источника данных сервера, на основе входящих параметров
	datasource, err := storage.NewDatasource(cfg.DatabaseDSN, cfg.StoreFile, cfg.StoreInterval, cfg.RestoreOnStart)
	if err != nil {
		cfg.ErrorLog.Fatal(err)
	}
	defer datasource.Close()

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   cfg.ErrorLog,  //	журнал ошибок приложения
		InfoLog:    cfg.InfoLog,   //	журнал информационных сообщений
		KeyToSign:  cfg.KeyToSign, //	ключ для подписи метрик по алгоритму HMAC c SHA256
		Datasource: datasource,    //	источник данных для хранения метрик
	}

	//	запускаем процесс слежение за сигналами на останов программы
	go termSignal()

	srv := &http.Server{
		Addr:     cfg.ServerAddress, //	адрес запуска сервера
		ErrorLog: app.ErrorLog,      //	журнал ошибок сервера
		Handler:  app.Routes(),      //	маршрутизатор сервера
	}
	//	запускаем сервер сбора метрик
	log.Fatal(srv.ListenAndServe())
}

// termSignal - функция слежения за сигналами на останов сервера
func termSignal() {
	// сигнальный канал для отслеживания системных вызовов на остановку программы
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//	запускаем слежение за сигнальным каналом
	for {
		sigTerm := <-signalChanel
		if sigTerm == syscall.SIGINT || sigTerm == syscall.SIGTERM || sigTerm == syscall.SIGQUIT {
			//	при получении сигнала, останавливаем программу с кодом - 0
			time.Sleep(1 * time.Second)
			log.Println("AGENT Gophermart SHUTDOWN (code 0)")
			os.Exit(0)
		}
	}
}
