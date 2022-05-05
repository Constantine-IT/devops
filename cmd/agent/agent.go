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

	var memStatistics runtime.MemStats //	экземпляр структуры для сохранения статистических данных

	pollCounter := &PollCounter{Count: 0} //	экземпляр структуры счётчика сбора метрик с mutex

	pollTicker := time.NewTicker(*PollInterval) //	тикер для выдачи сигналов на пересбор статистики
	time.Sleep(100 * time.Millisecond)
	reportTicker := time.NewTicker(*ReportInterval) //	тикер для выдачи сигнала на отправку статистики на сервер

	log.Println("AGENT - metrics collector STARTED with configuration:\n   ADDRESS: ", *ServerAddress, "\n   POLL_INTERVAL: ", *PollInterval, "\n   REPORT_INTERVAL: ", *ReportInterval, "\n   KEY for Signature: ", *KeyToSign)

	go func() { //	запускаем пересбор статистики раз в POLL_INTERVAL
		for {
			<-pollTicker.C
			//	считываем статиститку и увеличиваем счетчик считываний на 1
			runtime.ReadMemStats(&memStatistics)
			pollCounter.mutex.Lock() //	чуть позже заменим на атомарную операцию
			pollCounter.Count++
			pollCounter.mutex.Unlock()
		}
	}()
	go func() { //	запускаем отправку метрик на сервер раз в REPORT_INTERVAL
		for {
			<-reportTicker.C
			//	высылаем собранные метрики на сервер
			pollCounter.mutex.Lock()
			sendMetrics(memStatistics, pollCounter.Count, *ServerAddress, *KeyToSign)
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

/* for increment14
package main

import (
    "fmt"

    "github.com/shirou/gopsutil/v3/mem"
    // "github.com/shirou/gopsutil/mem"  // to use v2
)

func main() {
    v, _ := mem.VirtualMemory()

    // almost every return value is a struct
    fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

    // convert to JSON. String() is also implemented
    fmt.Println(v)
}
*/
