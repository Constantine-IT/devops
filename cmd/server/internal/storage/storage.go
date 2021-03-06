package storage

import (
	"errors"
	"io"
	"log"
)

//	Методы для работы с данными в структуре в оперативной памяти - Storage

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
	//	каждая запись - это структура Metrics - (NAME + Type + VALUE/DELTA)
	//	для метрик типа gauge задано только поле Value
	//	для метрик типа count задано только поле Delta

	if mType == "gauge" { //	для метрик типа GAUGE
		flgExist := 0 //	изначально предполагаем, что такой метрики у нас в базе нет

		for i, m := range s.Data { //	ищем метрику в базе
			if m.ID == name { //	если метрика уже существует, то для метрик типа GAUGE
				s.Data[i].Value = value //	новое значение заменяют предыдущее значение
				flgExist = 1            //	выставляем флаг, чтобы пропустить создание новой метрики
				break                   //	завершаем перебор строк хранилища
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

	if mType == "counter" { //	для метрик типа COUNT
		flgExist := 0 //	изначально предполагаем, что такой метрики у нас в базе нет

		for i, m := range s.Data { //	ищем метрику в базе
			if m.ID == name { //	если метрика уже существует, то для метрик типа COUNT
				s.Data[i].Delta += delta //	новое значение суммируется со старым
				flgExist = 1             //	выставляем флаг, чтобы пропустить создание новой метрики
				break                    //	завершаем перебор строк хранилища
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

	//	проверяем, есть ли запись с искомой метрикой в нашей базе
	for _, m := range s.Data {
		if m.ID == name {
			// если метрика с искомым имененм найдена, возвращаем её тип и значение, с флагом flag=1
			return m.MType, m.Delta, m.Value, 1
		}
	}
	return "", 0, 0, 0 //	если метрика с искомым имененм НЕ найдена, возвращаем flag=0
}

//	GetAll - метод возсвращает всё содержимое хранилища метрик
func (s *Storage) GetAll() (result []Metrics) {
	// блокируем хранилище на время считывания информации
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Data
}

//	Close - метод закрытия структур хранения
func (s *Storage) Close() { //	при остановке сервера
	if fileWriter != nil { //	если открыт файл-хранилище
		//	 сбрасываем содержимое структур оперативной памяти в файл
		DumpToFile(s)
		// закрываем writer для файла-хранилища
		fileWriter.Close()
	}
}

//	InitialFulfilment - метод первичного заполнения хранилища метрик из файла-хранилища, при старте сервера
func (s *Storage) InitialFulfilment() {
	//	блокируем хранилище в оперативной памяти на время заливки данных
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for {
		addMetrica := true                //	флаг, показывающий, будем ли добавлять метрику в хранилище
		metrica, err := fileReader.Read() //	считываем записи по одной из файла-хранилища

		if errors.Is(err, io.EOF) { //	когда дойдем до конца файла - выходим из цикла чтения
			log.Println("initial load metrics from file - SUCCESS")
			break
		}
		if err != nil { //	при ошибке чтения метрики - пропускаем её и читаем файл дальше
			log.Println("file read error due to InitialFulfilment process")
			continue
		}

		//	добавляем считанную метрику в хранилище в оперативной памяти - Storage
		for _, m := range s.Data {
			if m.ID == metrica.ID {
				addMetrica = false //	если метрика уже есть в хранилище, выставляем флаг на пропуск этой метрики
				break
			}
		}
		if addMetrica { //	метрики с флагом = true - добавялем в хранилище
			s.Data = append(s.Data, *metrica)
		}
	}
}
