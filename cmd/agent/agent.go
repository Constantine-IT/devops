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

	"github.com/shirou/gopsutil/v3/mem"
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

	var memStatistics runtime.MemStats      //	экземпляр структуры для сохранения статистических данных RUNTIME
	var GopStatistics mem.VirtualMemoryStat //	экземпляр структуры для сохранения статистических данных GOPSUTIL

	pollCounter := &PollCounter{Count: 0} //	экземпляр структуры счётчика сбора метрик с mutex

	pollTicker := time.NewTicker(*PollInterval) //	тикер для выдачи сигналов на пересбор статистики RUNTIME
	gopTicker := time.NewTicker(*PollInterval)  //	 //	тикер для выдачи сигналов на пересбор статистики GOPSUTIL
	time.Sleep(500 * time.Millisecond)
	reportTicker := time.NewTicker(*ReportInterval) //	тикер для выдачи сигнала на отправку статистики на сервер

	log.Println("AGENT - metrics collector STARTED with configuration:\n   ADDRESS: ", *ServerAddress, "\n   POLL_INTERVAL: ", *PollInterval, "\n   REPORT_INTERVAL: ", *ReportInterval, "\n   KEY for Signature: ", *KeyToSign)

	go func() { //	запускаем пересбор статистики RUNTIME раз в POLL_INTERVAL в отдельной горутине
		for {
			<-pollTicker.C
			//	считываем статиститку и увеличиваем счетчик считываний на 1
			pollCounter.mutex.Lock() //	чуть позже заменим на атомарную операцию
			pollCounter.Count++
			pollCounter.mutex.Unlock()
			runtime.ReadMemStats(&memStatistics)
		}
	}()

	go func() { //	запускаем пересбор статистики GOPSUTIL раз в POLL_INTERVAL в отдельной горутине
		for {
			<-gopTicker.C
			//	считываем статиститку и увеличиваем счетчик считываний на 1
			g, _ := mem.VirtualMemory()
			GopStatistics = *g

		}
	}()

	go func() { //	запускаем отправку метрик на сервер раз в REPORT_INTERVAL в отдельной горутине
		for {
			<-reportTicker.C
			//	высылаем собранные метрики на сервер
			pollCounter.mutex.Lock()
			sendMetrics(memStatistics, GopStatistics, pollCounter.Count, *ServerAddress, *KeyToSign)
			//	после передачи метрик, сбрасываем счетчик циклов измерения метрик
			pollCounter.Count = 0
			pollCounter.mutex.Unlock()
		}
	}()

	// создаём сигнальный канал для отслеживания системных вызовов на остановку агента
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//	запускаем процесс слежение за сигналами на останов агента
	for {
		s := <-signalChanel //	при получении сигнала на закрытие приложения - делаем os.Exit со статусом 0
		if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
			log.Println("AGENT metrics collector (code 0) SHUTDOWN")
			os.Exit(0)
		}
	}
}
