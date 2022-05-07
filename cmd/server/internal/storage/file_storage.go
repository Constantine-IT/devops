package storage

import (
	"encoding/json"
	"os"
	"sync"
)

//	Структуры и методы работы с файлом-хранилищем метрик

//	fileWriter и fileReader - рабочие экземпляры файловых дескрипторов чтения и записи
var fileWriter *writer
var fileReader *reader

//	writer - структура файлового дескриптора для записи
type writer struct {
	mutex   sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

//	NewWriter - конструктор, создающий экземпляр файлового дескриптора для записи
func NewWriter(fileName string) (*writer, error) {
	//	файл открывается только на запись с добавлением в конец файла, если файла нет - создаем
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777) //	|os.O_APPEND|os.O_TRUNC
	if err != nil {
		return nil, err
	}
	return &writer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Write - метод записи в файл для экземпляра файлового дескриптора для записи
func (p *writer) Write(metrica *Metrics) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.encoder.Encode(&metrica)
}

// Close - метод закрытия файла для экземпляра файлового дескриптора для записи
func (p *writer) Close() error {
	return p.file.Close()
}

//	reader - структура файлового дескриптора для чтения
type reader struct {
	mutex   sync.Mutex
	file    *os.File
	decoder *json.Decoder
}

//	NewReader - конструктор, создающий экземпляр файлового дескриптора для чтения
func NewReader(fileName string) (*reader, error) {
	//	файл открывается только на чтение, если файла нет - создаем
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// Read - метод чтения из файла для экземпляра файлового дескриптора для чтения
func (c *reader) Read() (*Metrics, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	metrica := &Metrics{}
	if err := c.decoder.Decode(&metrica); err != nil {
		return nil, err
	}
	return metrica, nil
}

// Close - метод закрытия файла для экземпляра файлового дескриптора для чтения
func (c *reader) Close() error {
	return c.file.Close()
}

//	DumpToFile - сбрасывает все метрики в файловое хранилище, затирая его содержимое новой информацией
func DumpToFile(s *Storage) error {
	//	перебираем все строки хранилища метрик в оперативной памяти по одной и вставляем в файл-хранилище
	for _, m := range s.Data {
		if err := fileWriter.Write(&m); err != nil {
			return err
		}
	}
	return nil
}
