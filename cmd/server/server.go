package main

import (
	"flag"
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
		*ServerAddress = u
	}
	if u, flg := os.LookupEnv("DATABASE_DSN"); flg { //	DATABASE_DSN — адрес подключения к БД (PostgreSQL)
		*DatabaseDSN = u
	}
	if u, flg := os.LookupEnv("STORE_FILE"); flg { //	STORE_FILE — путь до файла с сокращёнными метриками
		*StoreFile = u
	}
	if u, flg := os.LookupEnv("STORE_INTERVAL"); flg { //	STORE_INTERVAL — интервал сброса показания сервера на диск
		*StoreInterval, _ = time.ParseDuration(u) //	конвертируеим считанный string в интервал в секундах
	}
	if u, flg := os.LookupEnv("RESTORE"); flg { //	RESTORE — определяет, загружать ли метрики файла при старте сервера
		if u == "false" { //	если флаг равен FALSE, то присвоим переменной значение FALSE
			*RestoreOnStart = false
		} else { //	для всех иных явно заданных значений флага, присваиваем переменной значение TRUE
			*RestoreOnStart = true
		}
	}
	if u, flg := os.LookupEnv("KEY"); flg { //	KEY - ключ подписи передаваемых метрик
		*KeyToSign = u
	}

	if *StoreFile != "/tmp/devops-metrics-db.json" { //	для автотестов использующих файл, а не БД
		*DatabaseDSN = "" //	чтобы не возникало конфликтов с БД
	}
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)                  // logger для информационных сообщений
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) // logger для сообщений об ошибках

	//	конструктор источника данных сервера, на основе входящих параметров
	datasource, err := storage.NewDatasource(*DatabaseDSN, *StoreFile, *StoreInterval, *RestoreOnStart)
	if err != nil {
		errorLog.Fatal(err)
	}

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   errorLog,   //	журнал ошибок приложения
		InfoLog:    infoLog,    //	журнал информационных сообщений
		KeyToSign:  *KeyToSign, //	ключ для подписи метрик по алгоритму HMAC c SHA256
		Datasource: datasource, //	источник данных для хранения метрик
	}

	//	при остановке сервера отложенно закроем все источники данных
	defer app.Datasource.Close()

	srv := &http.Server{
		Addr:     *ServerAddress, //	адрес запуска сервера
		ErrorLog: app.ErrorLog,   //	журнал ошибок сервера
		Handler:  app.Routes(),   //	маршрутизатор сервера
	}
	go func() { //	запускаем сервер сбора метрик в отдельной горутине
		log.Println("SERVER - metrics collector STARTED with configuration:\n   ADDRESS: ", *ServerAddress, "\n   DATABASE_DSN: ", *DatabaseDSN, "\n   STORE_FILE: ", *StoreFile, "\n   STORE_INTERVAL: ", *StoreInterval, "\n   RESTORE: ", *RestoreOnStart, "\n   KEY for Signature: ", *KeyToSign)
		log.Fatal(srv.ListenAndServe())
	}()

	// создаём сигнальный канал для отслеживания системных вызовов на остановку сервера
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//	запускаем процесс слежение за сигналами на останов сервера
	for {
		s := <-signalChanel
		if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
			// в случае корректного останова сервера - закрываем все структуры хранения
			app.Datasource.Close()
			log.Println("SERVER metrics collector (code 0) SHUTDOWN")
			os.Exit(0)
		}
	}

}
