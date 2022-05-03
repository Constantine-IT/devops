package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type PollCounter struct {
	Count int64
	mutex sync.Mutex
}

func main() {

	var m runtime.MemStats

	pollCounter := &PollCounter{Count: 0}

	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "ADDRESS - адрес сервера-агрегатора метрик")
	PollInterval := flag.Duration("p", 2*time.Second, "POLL_INTERVAL - интервал обновления метрик (сек.)")
	ReportInterval := flag.Duration("r", 10*time.Second, "REPORT_INTERVAL - интервал отправки метрик на сервер (сне.)")
	//	парсим флаги
	flag.Parse()

	//	считываем переменные окружения
	//	если они заданы - переопределяем соответствующие локальные переменные:
	if aString, flg := os.LookupEnv("ADDRESS"); flg {
		*ServerAddress = aString
	}
	if pString, flg := os.LookupEnv("POLL_INTERVAL"); flg {
		pInt, _ := strconv.Atoi(pString)                  //	LookupEnv всегда считывает тип string - преобразуем его в int
		*PollInterval = time.Duration(pInt) * time.Second //	и зададим интервал в pInt секунд
	}
	if rString, flg := os.LookupEnv("REPORT_INTERVAL"); flg {
		rInt, _ := strconv.Atoi(rString)                    //	LookupEnv всегда считывает тип string - преобразуем его в int
		*ReportInterval = time.Duration(rInt) * time.Second //	и зададим интервал в rInt секунд

	}

	pollTicker := time.NewTicker(*PollInterval)
	time.Sleep(100 * time.Millisecond)
	reportTicker := time.NewTicker(*ReportInterval)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	log.Println("AGENT metrics collector START")
	for {
		select {
		case s := <-signalChanel:
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				log.Println("AGENT metrics collector (code 0) SHUTDOWN")
				os.Exit(0)
			}
		case <-pollTicker.C:
			//	считываем статиститку и увеличиваем счетчик считываний на 1
			runtime.ReadMemStats(&m)
			pollCounter.mutex.Lock()
			pollCounter.Count++
			pollCounter.mutex.Unlock()
			//log.Println("AGENT: Statistics renewed")

		case <-reportTicker.C:
			//	высылаем собраннуе метрики на сервер
			sendMetrics(&m, pollCounter, *ServerAddress)
			//log.Println("AGENT: Metrics sent to server")
		}
	}
}
