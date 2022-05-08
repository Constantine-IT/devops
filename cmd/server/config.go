package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	ServerAddress  string
	KeyToSign      string
	StoreFile      string
	StoreInterval  time.Duration
	RestoreOnStart bool
	DatabaseDSN    string
	InfoLog        *log.Logger
	ErrorLog       *log.Logger
}

//	newConfig - функция-конфигуратор приложения через считывание флагов и переменных окружения
func newConfig() (cfg Config) {
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

	//	DATABASE_DSN имеет приоритет над FILE_STORE, то есть использование базы данных отменяет запись метрик в файл,
	//	но возможно считывание метрик из файла при старте сервера, при наличии RESTORE=true

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
	if u, flg := os.LookupEnv("STORE_INTERVAL"); flg { //	STORE_INTERVAL — интервал сохранения метрик на диск
		*StoreInterval, _ = time.ParseDuration(u) //	конвертируем считанный string во временной интервал
	}
	if u, flg := os.LookupEnv("RESTORE"); flg { //	RESTORE — определяет, загружать ли метрики файла при старте сервера
		if u == "false" { //	если флаг равен FALSE, то присвоим переменной значение FALSE
			*RestoreOnStart = false
		} else { //	для всех иных ЯВНО заданных значений флага, присваиваем переменной значение TRUE (значение по умолчанию)
			*RestoreOnStart = true
		}
	}
	if u, flg := os.LookupEnv("KEY"); flg { //	KEY - ключ подписи передаваемых метрик
		*KeyToSign = u
	}
	// logger для информационных сообщений и для сообщений об ошибках
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// сигнальный канал для отслеживания системных вызовов на остановку сервера
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//	запускаем процесс слежение за сигналами на останов сервера
	go func() {
		for {
			s := <-signalChanel
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				cfg.InfoLog.Println("SERVER metrics collector normal SHUTDOWN (code 0)")
				os.Exit(0)
			}
		}
	}()

	//	собираем конфигурацию сервера
	cfg = Config{
		ServerAddress:  *ServerAddress,
		KeyToSign:      *KeyToSign,
		StoreFile:      *StoreFile,
		StoreInterval:  *StoreInterval,
		RestoreOnStart: *RestoreOnStart,
		DatabaseDSN:    *DatabaseDSN,
		InfoLog:        infoLog,
		ErrorLog:       errorLog,
	}

	//	выводим в лог конфигурацию сервера
	log.Println("SERVER - metrics collector STARTED with configuration:\n   ADDRESS: ", cfg.ServerAddress, "\n   DATABASE_DSN: ", cfg.DatabaseDSN, "\n   STORE_FILE: ", cfg.StoreFile, "\n   STORE_INTERVAL: ", cfg.StoreInterval, "\n   RESTORE: ", cfg.RestoreOnStart, "\n   KEY for Signature: ", cfg.KeyToSign)

	return cfg
}
