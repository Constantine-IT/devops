package main

import (
	"flag"
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

	var m runtime.MemStats

	const (
		pollInterval   = 2 * time.Second
		reportInterval = 10 * time.Second
	)

	pollCounter := &PollCounter{Count: 0}

	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "SERVER_ADDRESS - адрес сервера-агрегатора метрик")
	//	парсим флаги
	flag.Parse()

	//log.Println("AGENT: metrics collector start")

	pollTicker := time.NewTicker(pollInterval)
	time.Sleep(100 * time.Millisecond)
	reportTicker := time.NewTicker(reportInterval)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	for {
		select {
		case s := <-signalChanel:
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				//log.Println("AGENT metrics collector shutdown normal")
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
