package main

import (
	"flag"
	"log"
	"os"
	"time"
)

type Config struct {
	ServerAddress  string        //	адрес сервера-агрегатора метрик
	KeyToSign      string        //	ключ подписи передаваемых метрик
	PollInterval   time.Duration //	интервал обновления метрик
	ReportInterval time.Duration //	интервал отправки метрик на сервер
	InfoLog        *log.Logger   //	logger для информационных сообщений
	ErrorLog       *log.Logger   //	logger для сообщений об ошибках
}

//	newConfig - функция-конфигуратор приложения через считывание флагов и переменных окружения
func newConfig() (cfg Config) {
	//	Приоритеты настроек:
	//	1.	Переменные окружения - ENV
	//	2.	Значения, задаваемые флагами при запуске из консоли
	//	3.	Значения по умолчанию.
	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS - адрес сервера-агрегатора метрик")
	KeyToSign := flag.String("k", "", "KEY - ключ подписи передаваемых метрик")
	PollInterval := flag.Duration("p", 2*time.Second, "POLL_INTERVAL - интервал обновления метрик (сек.)")
	ReportInterval := flag.Duration("r", 10*time.Second, "REPORT_INTERVAL - интервал отправки метрик на сервер (сне.)")
	//	парсим флаги
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if addrString, flg := os.LookupEnv("ADDRESS"); flg {
		*ServerAddress = addrString
	}
	if keyString, flg := os.LookupEnv("KEY"); flg {
		*KeyToSign = keyString
	}
	if pollString, flg := os.LookupEnv("POLL_INTERVAL"); flg {
		*PollInterval, _ = time.ParseDuration(pollString) //	конвертируеим считанный string в интервал в секундах
	}
	if reportString, flg := os.LookupEnv("REPORT_INTERVAL"); flg {
		*ReportInterval, _ = time.ParseDuration(reportString) //	конвертируеим считанный string в интервал в секундах
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)                  // logger для информационных сообщений
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) // logger для сообщений об ошибках

	//	собираем конфигурацию агента
	cfg = Config{
		ServerAddress:  *ServerAddress,
		KeyToSign:      *KeyToSign,
		PollInterval:   *PollInterval,
		ReportInterval: *ReportInterval,
		InfoLog:        infoLog,
		ErrorLog:       errorLog,
	}

	//	выводим в лог конфигурацию агента
	log.Println("AGENT - metrics collector STARTED with configuration:\n   ADDRESS: ", cfg.ServerAddress, "\n   POLL_INTERVAL: ", cfg.PollInterval, "\n   REPORT_INTERVAL: ", cfg.ReportInterval, "\n   KEY for Signature: ", cfg.KeyToSign)

	return cfg
}
