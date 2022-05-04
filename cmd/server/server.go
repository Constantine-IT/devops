package main

import (
	"flag"
	"github.com/Constantine-IT/devops/cmd/server/internal/handlers"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//	Приоритеты настроек:
	//	1.	Переменные окружения - ENV
	//	2.	Значения, задаваемые флагами при запуске из консоли
	//	3.	Значения по умолчанию.

	//	Считываем и парсим флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS — адрес запуска HTTP-сервера")
	KeyToSign := flag.String("k", "", "KEY - ключ подписи передаваемых метрик")
	StoreFile := flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE — путь до файла с сокращёнными метриками")
	StoreInterval := flag.Duration("i", 300*time.Second, "STORE_INTERVAL — интервал сброса показания сервера на диск")
	RestoreOnStart := flag.Bool("r", true, "RESTORE — определяет, загружать ли метрики файла при старте сервера")
	DatabaseDSN := flag.String("d", "", "DATABASE_DSN — адрес подключения к БД (PostgreSQL)")
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if u, flg := os.LookupEnv("ADDRESS"); flg { //	ADDRESS — адрес запуска HTTP-сервера
		log.Println("ENV:   ADDRESS set to: ", u)
		*ServerAddress = u
	}
	if u, flg := os.LookupEnv("KEY"); flg { //	KEY - ключ подписи передаваемых метрик
		*KeyToSign = u
	}
	if u, flg := os.LookupEnv("STORE_FILE"); flg { //	STORE_FILE — путь до файла с сокращёнными метриками
		log.Println("ENV:   STORE_FILE set to: ", u)
		*StoreFile = u
	}
	if u, flg := os.LookupEnv("STORE_INTERVAL"); flg { //	STORE_INTERVAL — интервал сброса показания сервера на диск
		log.Println("ENV:   STORE_INTERVAL set to: ", u)
		*StoreInterval, _ = time.ParseDuration(u) //	конвертируеим считанный string в интервал в секундах
	}
	if u, flg := os.LookupEnv("RESTORE"); flg { //	RESTORE — определяет, загружать ли метрики файла при старте сервера
		if u == "false" { //	если флаг равен FALSE, то присвоим переменной значение FALSE
			log.Println("ENV:   RESTORE set to FALSE")
			*RestoreOnStart = false
		} else { //	для всех иных явно заданных значений флага, присваиваем переменной значение TRUE
			log.Println("ENV:   RESTORE set to TRUE")
			*RestoreOnStart = true
		}
	}
	if u, flg := os.LookupEnv("DATABASE_DSN"); flg { //	DATABASE_DSN — адрес подключения к БД (PostgreSQL)
		log.Println("ENV:   DATABASE_DSN set to: ", u)
		*DatabaseDSN = u
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)                  // logger для информационных сообщений
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) // logger для сообщений об ошибках

	datasource, err := storage.NewDatasource(*DatabaseDSN, *StoreFile, *RestoreOnStart) // источник данных нашего приложения
	if err != nil {
		errorLog.Fatal(err)
	}

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   errorLog,   //	журнал ошибок
		InfoLog:    infoLog,    //	журнал информационных сообщений
		KeyToSign:  *KeyToSign, //	ключ для подписи метрик по алгоритму SHA256
		Datasource: datasource, //	источник данных для хранения URL
	}

	//	при остановке сервера отложенно закроем все источники данных
	defer app.Datasource.Close()

	srv := &http.Server{
		Addr:     *ServerAddress, //	адрес запуска сервера
		ErrorLog: app.ErrorLog,   //	журнал ошибок сервера
		Handler:  app.Routes(),   //	маршрутизатор сервера
	}
	go func() { //	запуск сервера с конфигурацией srv
		log.Println("SERVER - metrics collector STARTED with configuration:\n   ADDRESS: ", *ServerAddress, "\n   DATABASE_DSN: ", *DatabaseDSN, "\n   STORE_FILE: ", *StoreFile, "\n   STORE_INTERVAL: ", *StoreInterval, "\n   RESTORE: ", *RestoreOnStart, "\n   KEY for Signature: ", *KeyToSign)
		log.Fatal(srv.ListenAndServe())
	}()

	if *StoreInterval <= 0 {
		*StoreInterval = 1
	}
	// создаём тикер, подающий раз в StoreInterval секунд, сигнал на запись метрик в файл
	fileWriteTicker := time.NewTicker(*StoreInterval * time.Second)

	// создаём сигнальный канал для отслеживания системных сигналов на остановку сервера
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// запускаем слежение за каналами тикера записи в файл и сигналов на останов сервера
	for {
		select {
		case s := <-signalChanel:
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				//	перед завершением работы сервера - сохраняем все метрики в файл
				_ = app.Datasource.DumpToFile()
				log.Println("All metrics were written to file:", *StoreFile)
				log.Println("SERVER metrics collector (code 0) SHUTDOWN")
				os.Exit(0)
			}
		case <-fileWriteTicker.C:
			if *StoreFile != "" {
				//	пишем метрики в файл
				log.Println("All metrics were written to file:", *StoreFile)
				if err := app.Datasource.DumpToFile(); err != nil {
					log.Println("SERVER metrics collector unable to write to file - (code 1) SHUTDOWN")
					os.Exit(1)
				}
			}
		}
	}
}
