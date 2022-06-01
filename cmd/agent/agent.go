package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	s "sync/atomic"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/Constantine-IT/devops/cmd/agent/internal"
)

func main() {
	//	конфигурация приложения через считывание флагов и переменных окружения
	cfg := newConfig()

	var memStatistics runtime.MemStats      //	экземпляр структуры для сохранения статистических данных RUNTIME
	var GopStatistics mem.VirtualMemoryStat //	экземпляр структуры для сохранения статистических данных GOPSUTIL

	var pollCounter int64 = 0 //	счётчик циклов обновления статистики с атомарным управлением

	pollTicker := time.NewTicker(cfg.PollInterval)     //	тикер для выдачи сигналов на пересбор статистики RUNTIME
	gopTicker := time.NewTicker(cfg.PollInterval)      //	тикер для выдачи сигналов на пересбор статистики GOPSUTIL
	time.Sleep(cfg.PollInterval / 2)                   //	вводим задержку, чтобы сбор статистики не наложился на отправку на сервер
	reportTicker := time.NewTicker(cfg.ReportInterval) //	тикер для выдачи сигнала на отправку статистики на сервер
	defer pollTicker.Stop()
	defer gopTicker.Stop()
	defer reportTicker.Stop()

	//	создаем контекст для остановки служебных процессов по сигналу
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//	учёт запущенных горутин
	wg := &sync.WaitGroup{}
	//	добавляем 3 горутины: 2 на сбор статистики и 1 на отправку метрик на сервер
	wg.Add(3)

	go func() { //	запускаем сбор статистики RUNTIME раз в POLL_INTERVAL в отдельной горутине
		defer wg.Done()
		for {
			select {
			case <-pollTicker.C:
				//	считываем статистику и увеличиваем счетчик повторного сбора статистики на 1
				runtime.ReadMemStats(&memStatistics)
				s.AddInt64(&pollCounter, 1)
			case <-ctx.Done(): //	при подаче сигнала на останов сервера, прерываем сбор статистики
				cfg.InfoLog.Println("RUNTIME statistics collector has stopped")
				return
			}
		}
	}()

	go func() { //	запускаем пересбор статистики GOPSUTIL раз в POLL_INTERVAL в отдельной горутине
		defer wg.Done()
		for {
			select {
			case <-pollTicker.C:
				g, _ := mem.VirtualMemory()
				GopStatistics = *g
			case <-ctx.Done(): //	при подаче сигнала на останов сервера, прерываем сбор статистики
				cfg.InfoLog.Println("GOPSUTIL statistics collector has stopped")
				return
			}
		}
	}()

	go func() { //	запускаем отправку метрик на сервер раз в REPORT_INTERVAL в отдельной горутине
		defer wg.Done()
		for {
			select {
			case <-pollTicker.C:
				//	высылаем собранные метрики на сервер
				internal.SendMetrics(memStatistics, GopStatistics, pollCounter, cfg.ServerAddress, cfg.KeyToSign)
				//	после передачи метрик, сбрасываем счетчик циклов измерения метрик в значение = 0
				s.StoreInt64(&pollCounter, 0)
			case <-ctx.Done(): //	при подаче сигнала на останов сервера, прерываем отправку статистики на сервер
				cfg.InfoLog.Println("Statistics sender has stopped")
				return
			}
		}
	}()

	//	запускаем процесс слежение за сигналами на останов программы
	go termSignal(cancel)

	//	ждём до закрытия всех горутин
	wg.Wait()
}

// termSignal - функция слежения за сигналами на остановку программы
func termSignal(cancel context.CancelFunc) {
	// сигнальный канал для отслеживания системных вызовов на остановку программы
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//	запускаем слежение за сигнальным каналом
	for {
		sigTerm := <-signalChanel
		if sigTerm == syscall.SIGINT || sigTerm == syscall.SIGTERM || sigTerm == syscall.SIGQUIT {
			cancel() //	при получении сигнала, прерываем контекст и останавливаем программу с кодом - 0
			time.Sleep(1 * time.Second)
			log.Println("AGENT Gophermart SHUTDOWN (code 0)")
			os.Exit(0)
		}
	}
}
