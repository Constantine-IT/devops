package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {

	var pollCount uint64
	var m runtime.MemStats

	const (
		pollInterval   = 2 * time.Second
		reportInterval = 10 * time.Second
	)

	//	Считываем флаги запуска из командной строки и задаём значения по умолчанию, если флаг при запуске не указан
	ServerAddress := flag.String("a", "127.0.0.1:8080", "SERVER_ADDRESS - адрес сервера-агрегатора метрик")
	//	парсим флаги
	flag.Parse()

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	log.Println("AGENT metrics collector start")

	for {
		select {
		case s := <-signalChanel:
			if s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGQUIT {
				log.Println("AGENT metrics collector shutdown normal")
				os.Exit(1)
			}
		case <-pollTicker.C:
			runtime.ReadMemStats(&m)
			pollCount++

		case <-reportTicker.C:
			sendMetrica(m, pollCount, *ServerAddress)
		}
	}
}
