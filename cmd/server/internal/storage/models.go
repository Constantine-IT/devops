package storage

import (
	"database/sql"
	"errors"
	"sync"
)

//	Datasource - интерфейс источника данных для метрик
//	может реализовываться базой данных (Database) или структурами в оперативной памяти (Storage)
type Datasource interface {
	Insert(name, mType string, delta int64, value float64) error
	Get(name string) (mType string, delta int64, value float64, flg int)
	GetAll() (result []Metrics)
	Close()
}

//	Database - структура хранилища метрик, обертывающая пул подключений к базе данных
type Database struct {
	DB *sql.DB
}

//	Metrics - структура для хранения метрик в оперативной памяти
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

//	ErrEmptyNotAllowed - ошибка возникающая при попытке вставить пустое значение в любое поле структуры хранения URL
var ErrEmptyNotAllowed = errors.New("DataBase: empty value is not allowed")
