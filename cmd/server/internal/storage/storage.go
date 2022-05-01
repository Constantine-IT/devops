package storage

import (
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
	return s.Data[name].mType, value, 1 //	если метрика NAME найдена, возвращаем flag=1
}

func (s *Storage) Close() {
}
