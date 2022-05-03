package main

import (
	"flag"
	"github.com/Constantine-IT/devops/cmd/server/internal/handlers"
	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
	"log"
	"net/http"
	"os"
)

func main() {
	//	Приоритеты настроек:
	//	1.	Переменные окружения - ENV
	//	2.	Значения, задаваемые флагами при запуске из консоли
	//	3.	Значения по умолчанию.

	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS - адрес запуска HTTP-сервера")
	BaseURL := flag.String("b", "http://127.0.0.1:8080", "BASE_URL - базовый адрес результирующего сокращённого URL")
	DatabaseDSN := flag.String("d", "", "DATABASE_DSN - адрес подключения к БД (PostgreSQL)")
	FileStorage := flag.String("f", "", "FILE_STORAGE_PATH - путь до файла с сокращёнными URL")
	//	парсим флаги
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if u, flg := os.LookupEnv("ADDRESS"); flg {
		*ServerAddress = u
	}
	if u, flg := os.LookupEnv("BASE_URL"); flg {
		*BaseURL = u
	}
	if u, flg := os.LookupEnv("DATABASE_DSN"); flg {
		*DatabaseDSN = u
	}
	if u, flg := os.LookupEnv("FILE_STORAGE_PATH"); flg {
		*FileStorage = u
	}

	//	инициализируем logger для информационных сообщений
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	//	инициализируем logger для сообщений об ошибках
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	//	инициализируем источники данных нашего приложения для работы с URL
	datasource, err := storage.NewDatasource(*DatabaseDSN, *FileStorage)
	if err != nil {
		errorLog.Fatal(err)
	}

	//	инициализируем контекст нашего приложения
	app := &handlers.Application{
		ErrorLog:   errorLog,   //	журнал ошибок
		InfoLog:    infoLog,    //	журнал информационных сообщений
		BaseURL:    *BaseURL,   //	базоовый адрес сервера
		Datasource: datasource, //	источник данных для хранения URL
	}

	//	при остановке сервера отложенно закроем все источники данных
	defer app.Datasource.Close()

	//	запуск сервера
	infoLog.Printf("Server started at address: %s", *ServerAddress)

	srv := &http.Server{
		Addr:     *ServerAddress,
		ErrorLog: app.ErrorLog,
		Handler:  app.Routes(),
	}
	syserr := srv.ListenAndServe()
	if syserr != nil {
		os.Exit(0)
	}
}
