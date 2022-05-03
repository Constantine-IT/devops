package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type PollCounter struct {
	Count int64
	mutex sync.Mutex
}

func main() {
	//	Приоритеты настроек:
	//	1.	Переменные окружения - ENV
	//	2.	Значения, задаваемые флагами при запуске из консоли
	//	3.	Значения по умолчанию.
	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS - адрес сервера-агрегатора метрик")
	PollInterval := flag.Duration("p", 2*time.Second, "POLL_INTERVAL - интервал обновления метрик (сек.)")
	ReportInterval := flag.Duration("r", 10*time.Second, "REPORT_INTERVAL - интервал отправки метрик на сервер (сне.)")
	//	парсим флаги
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if addrString, flg := os.LookupEnv("ADDRESS"); flg {
		*ServerAddress = addrString
	}
	if pollString, flg := os.LookupEnv("POLL_INTERVAL"); flg {
		*PollInterval, _ = time.ParseDuration(pollString) //	конвертируеим считанный string в интервал в секундах
	}
	if reportString, flg := os.LookupEnv("REPORT_INTERVAL"); flg {
		*ReportInterval, _ = time.ParseDuration(reportString) //	конвертируеим считанный string в интервал в секундах
	}

	var memStatistics runtime.MemStats //	экземпляр структуры для сохранения статистических данных

	pollCounter := &PollCounter{Count: 0} //	экземпляр структуры счётчика сбора метрик с mutex

	pollTicker := time.NewTicker(*PollInterval) //	тикер для выдачи сигналов на пересбор статистики
	time.Sleep(100 * time.Millisecond)
	reportTicker := time.NewTicker(*ReportInterval) //	тикер для выдачи сигнала на отправку статистики на сервер

	//	сиоздаём сигнальный канал для отслеживания системных команд на закрытие приложения
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	log.Println("AGENT metrics collector START")

	for { // отслеживаем сигналы от созданных тикеров и сигнальных каналов, и реагируем на них соответственно
		select {
		case s := <-signalChanel: //	при получении сигнала на закрытие приложения - делаем os.Exit со статусом 0
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				log.Println("AGENT metrics collector (code 0) SHUTDOWN")
				os.Exit(0)
			}
		case <-pollTicker.C: //	запускаем пересбор статистики
			//	считываем статиститку и увеличиваем счетчик считываний на 1
			runtime.ReadMemStats(&memStatistics)
			pollCounter.mutex.Lock()
			pollCounter.Count++
			pollCounter.mutex.Unlock()

		case <-reportTicker.C: //	запускаем отправку метрик на сервер
			//	высылаем собраннуе метрики на сервер
			sendMetrics(&memStatistics, pollCounter, *ServerAddress)
		}
	}
}
