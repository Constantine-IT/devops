package storage

import (
	"sync"
)

//	Metrics - структура для хранения метрик и обмена данными между сервером и агентом
type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

//	Storage - структура хранилища метрик для работы в оперативной памяти
type Storage struct {
	Data  []Metrics
	mutex sync.Mutex
}

// Insert - Метод для сохранения метрик
func (s *Storage) Insert(name, mType string, delta int64, value float64) error {
	//	пустые значения к вставке в хранилище не допускаются
	if name == "" || mType == "" {
		return ErrEmptyNotAllowed
	}
	//	Блокируем структуру храниения в оперативной памяти на время записи информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	//	сохраняем метрики в оперативной памяти в структуре Storage
	//	каждая запись - это сопоставленная с NAME структура из (MetricaType + VALUE/DELTA) - Metrics

	if mType == "gauge" { //	для метрик типа GAUGE повторные вставки затирают предыдущие значения
		flgExist := 0
		for i, m := range s.Data {
			if m.ID == name {
				s.Data[i].Value = value
				flgExist = 1
			}
		}
		if flgExist == 0 {
			m := Metrics{
				ID:    name,
				MType: "gauge",
				Value: value,
			}
			s.Data = append(s.Data, m)
		}
	}

	if mType == "counter" { //	для метрик типа COUNT повторные вставки НЕ затирают предыдущие значения
		flgExist := 0
		for i, m := range s.Data {
			if m.ID == name {
				s.Data[i].Delta += delta //	новое значение суммируется со старым, содержащимся в базе
				flgExist = 1
			}
		}
		if flgExist == 0 {
			m := Metrics{
				ID:    name,
				MType: "counter",
				Delta: delta,
			}
			s.Data = append(s.Data, m)
		}
	}
	return nil
}

// Get - метод для нахождения значения метрики
func (s *Storage) Get(name string) (mType string, delta int64, value float64, flg int) {
	// блокируем хранилище на время считывания информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	//	проверяем, есть ли запись с запрашиваемым NAME в базе

	for _, m := range s.Data {
		if m.ID == name {
			// если метрика с искомым имененм найдена возвращаем её тип и значение, с флагом flag=1
			return m.MType, m.Delta, m.Value, 1
		}
	}
	return "", 0, 0, 0 //	если метрика с искомым имененм НЕ найдена, возвращаем flag=0
}

func (s *Storage) GetAll() []Metrics {
	// блокируем хранилище на время считывания информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Data
}

func (s *Storage) Close() {
}
