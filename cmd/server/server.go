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

	//STORE_INTERVAL (по умолчанию 300) — интервал времени в секундах,
	//по истечении которого текущие показания сервера сбрасываются на диск
	//(значение 0 — делает запись синхронной).

	//STORE_FILE (по умолчанию "/tmp/devops-metrics-db.json") — строка, имя файла, где хранятся значения
	//(пустое значение — отключает функцию записи на диск).

	//RESTORE (по умолчанию true) — булево значение (true/false),
	//определяющее, загружать или нет начальные значения из указанного файла при старте сервера.

	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS — адрес запуска HTTP-сервера")
	StoreFile := flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE — путь до файла с сокращёнными метриками")
	StoreInterval := flag.Duration("i", 300*time.Second, "STORE_INTERVAL — интервал сброса показания сервера на диск")
	RestoreOnStart := flag.Bool("r", true, "RESTORE — определяет, загружать ли метрики файла при старте сервера")
	DatabaseDSN := flag.String("d", "", "DATABASE_DSN — адрес подключения к БД (PostgreSQL)")
	//	парсим флаги
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if u, flg := os.LookupEnv("ADDRESS"); flg {
		log.Println("ENV:   ADDRESS set to: ", u)
		*ServerAddress = u
	}
	if u, flg := os.LookupEnv("STORE_FILE"); flg {
		log.Println("ENV:   STORE_FILE set to: ", u)
		*StoreFile = u
	}
	/*
		if u, flg := os.LookupEnv("STORE_INTERVAL"); flg {
			if strIntrvl, err := strconv.Atoi(u); err != nil { //	конвертируем считанный string в int
				log.Println("ENV:   error with parsing STORE_INTERVAL")
			} else {
				*StoreInterval = strIntrvl
			}
		}
	*/
	if u, flg := os.LookupEnv("STORE_INTERVAL"); flg {
		log.Println("ENV:   STORE_INTERVAL set to: ", u)
		*StoreInterval, _ = time.ParseDuration(u) //	конвертируеим считанный string в интервал в секундах
	}
	if u, flg := os.LookupEnv("RESTORE"); flg {
		if u == "false" { //	если флаг равен FALSE, то присвоим переменной значение FALSE
			log.Println("ENV:   RESTORE set to FALSE")
			*RestoreOnStart = false
		} else { //	для всех иных явно заданных значений флага, присваиваем переменной значение TRUE
			log.Println("ENV:   RESTORE set to TRUE")
			*RestoreOnStart = true
		}
	}
	if u, flg := os.LookupEnv("DATABASE_DSN"); flg {
		log.Println("ENV:   DATABASE_DSN set to: ", u)
		*DatabaseDSN = u
	}

	//	инициализируем logger для информационных сообщений
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	//	инициализируем logger для сообщений об ошибках
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	//	инициализируем источники данных нашего приложения для работы с URL
	datasource, err := storage.NewDatasource(*DatabaseDSN, *StoreFile, *RestoreOnStart)
	if err != nil {
		errorLog.Fatal(err)
	}

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   errorLog,   //	журнал ошибок
		InfoLog:    infoLog,    //	журнал информационных сообщений
		Datasource: datasource, //	источник данных для хранения URL
	}

	//	при остановке сервера отложенно закроем все источники данных
	defer app.Datasource.Close()

	srv := &http.Server{
		Addr:     *ServerAddress,
		ErrorLog: app.ErrorLog,
		Handler:  app.Routes(),
	}
	go func() { //	запуск сервера с конфигурацией srv
		log.Println("SERVER configuration. \n   ADDRESS: ", *ServerAddress, "\n   STORE_FILE: ", *StoreFile, "\n   STORE_INTERVAL: ", *StoreInterval, "\n   RESTORE: ", *RestoreOnStart)
		log.Fatal(srv.ListenAndServe())
	}()
	infoLog.Printf("Server started at address: %s", *ServerAddress)

	// создаём тикер, подающий раз в StoreInterval секунд, сигнал на запись метрик в файл
	if *StoreInterval <= 0 {
		*StoreInterval = 1
	}
	fileWriteInterval := time.Duration(*StoreInterval)
	fileWriteTicker := time.NewTicker(fileWriteInterval * time.Second)

	// создаём сигнальный канал для лтслеживания системных сигналов на остановку сервера
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
