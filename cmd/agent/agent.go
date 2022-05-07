package main

import (
	"github.com/Constantine-IT/devops/cmd/agent/internal"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

type PollCounter struct {
	Count int64
	mutex sync.Mutex
}

func main() {
	//	конфигурация приложения через считывание флагов и переменных окружения
	cfg := newConfig()

	var memStatistics runtime.MemStats      //	экземпляр структуры для сохранения статистических данных RUNTIME
	var GopStatistics mem.VirtualMemoryStat //	экземпляр структуры для сохранения статистических данных GOPSUTIL
	pollCounter := &PollCounter{Count: 0}   //	экземпляр структуры счётчика сбора метрик с mutex

	pollTicker := time.NewTicker(cfg.PollInterval)     //	тикер для выдачи сигналов на пересбор статистики RUNTIME
	gopTicker := time.NewTicker(cfg.PollInterval)      //	тикер для выдачи сигналов на пересбор статистики GOPSUTIL
	time.Sleep(cfg.PollInterval / 2)                   //	вводим задержку, чтобы сбор статистики не наложился на отправку на сервер
	reportTicker := time.NewTicker(cfg.ReportInterval) //	тикер для выдачи сигнала на отправку статистики на сервер
	defer pollTicker.Stop()
	defer gopTicker.Stop()
	defer reportTicker.Stop()

	//	учёт запущенных горутин
	wg := &sync.WaitGroup{}
	//	добавялем 3 горутины: 2 на сбор статистики и 1 на отправку метрик на сервер
	wg.Add(3)

	go func() { //	запускаем пересбор статистики RUNTIME раз в POLL_INTERVAL в отдельной горутине
		defer wg.Done()
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
		defer wg.Done()
		for {
			<-gopTicker.C
			g, _ := mem.VirtualMemory()
			GopStatistics = *g

		}
	}()

	go func() { //	запускаем отправку метрик на сервер раз в REPORT_INTERVAL в отдельной горутине
		defer wg.Done()
		for {
			<-reportTicker.C
			//	высылаем собранные метрики на сервер
			pollCounter.mutex.Lock()
			internal.SendMetrics(memStatistics, GopStatistics, pollCounter.Count, cfg.ServerAddress, cfg.KeyToSign)
			//	после передачи метрик, сбрасываем счетчик циклов измерения метрик
			pollCounter.Count = 0
			pollCounter.mutex.Unlock()
		}
	}()

	//	ждём до закрытия всех горутин
	wg.Wait()
}
