package storage

import (
	"sync"
)

//	Metrics - структура для хранения метрик
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

	if mType == "gauge" { //	для метрик типа GAUGE повторные вставки заменяют предыдущие значения
		flgExist := 0 //	изначально предполагаем, что такой метрики у нас в базе нет

		for i, m := range s.Data { //	ищем метрику в базе
			if m.ID == name { //	если метрика уже существует в базе, то для метрик типа GAUGE
				s.Data[i].Value = value //	новое значение заменяют предыдущие значения
				flgExist = 1            //	выставляем флаг, чтобы пропустить создание новой метрики
			}
		}
		if flgExist == 0 { //	если метрики в базе не существует, то создаем для неё новую запись
			m := Metrics{
				ID:    name,
				MType: "gauge",
				Delta: 0,
				Value: value,
			}
			s.Data = append(s.Data, m) //	добавляем созданную новую запись в базу
		}
	}

	if mType == "counter" { //	для метрик типа COUNT повторные вставки НЕ затирают предыдущие значения
		flgExist := 0 //	изначально предполагаем, что такой метрики у нас в базе нет

		for i, m := range s.Data { //	ищем метрику в базе
			if m.ID == name { //	если метрика уже существует в базе, то для метрик типа COUNT
				s.Data[i].Delta += delta //	новое значение суммируется со старым
				flgExist = 1             //	выставляем флаг, чтобы пропустить создание новой метрики
			}
		}
		if flgExist == 0 { //	если метрики в базе не существует, то создаем для неё новую запись
			m := Metrics{
				ID:    name,
				MType: "counter",
				Delta: delta,
				Value: 0,
			}
			s.Data = append(s.Data, m) //	добавляем созданную новую запись в базу
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
			//log.Println("GET return:", name, mType, delta, value, flg)
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

func (s *Storage) Close() error {
	//	при остановке сервера закрываем reader и writer для файла-хранилища URL
	if err := fileReader.Close(); err != nil {
		return err
	}
	if err := fileWriter.Close(); err != nil {
		return err
	}
	return nil
}

//	DumpToFile - сбрасывает все метрики в файловое хранилище
func (s *Storage) DumpToFile() error {
	//	перебираем все строки хранилища метрик в оперативной памяти по одной и вставляем в файл-хранилище
	for _, metrica := range s.Data {
		if err := fileWriter.Write(&metrica); err != nil {
			return err
		}
	}
	return nil
}
