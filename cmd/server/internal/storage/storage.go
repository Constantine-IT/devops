package storage

import (
	"log"
	"sync"
)

//	Storage - структура хранилища метрик для работы в оперативной памяти
type Storage struct {
	Data  map[string]MetricaRow
	mutex sync.Mutex
}

// Insert - Метод для сохранения метрик
func (s *Storage) Insert(name, mType, value string) error {
	//	пустые значения к вставке в хранилище не допускаются
	if name == "" || mType == "" || value == "" {
		return ErrEmptyNotAllowed
	}
	//	Блокируем структуру храниения в оперативной памяти на время записи информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	//	сохраняем метрики в оперативной памяти в структуре Storage
	//	каждая запись - это сопоставленная с NAME структура из (MetricaType + VALUE) - MetricaRow
	s.Data[name] = MetricaRow{mType, value}
	//log.Println(name, s.Data[name])
	return nil
}

// Get - метод для нахождения значения метрики
func (s *Storage) Get(name string) (mType, value string, flg int) {
	// блокируем хранилище на время считывания информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	//	проверяем, есть ли запись с запрашиваемым NAME в базе
	if _, ok := s.Data[name]; !ok {
		return "", "", 0
	} //	если метрика NAME не найдена, возвращаем flag=0
	log.Println(name, s.Data[name])
	return s.Data[name].mType, s.Data[name].value, 1 //	если метрика NAME найдена, возвращаем flag=1
}

func (s *Storage) GetAll() ([]MetricaValue, bool) {
	// блокируем хранилище на время считывания информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	metricaValues := make([]MetricaValue, 0)

	for name, metricaRows := range s.Data {
		metricaValues = append(metricaValues, MetricaValue{name, metricaRows.value})
	}
	//log.Println(metricaValues)
	if len(metricaValues) == 0 { //	если записей не найдено - выставляем FLAG в положение FALSE
		return nil, false
	} else {
		return metricaValues, true //	если записей найдены - выставляем FLAG в положение TRUE и возвращаем их
	}
}

func (s *Storage) Close() {
}
