package storage

import "errors"

//	Datasource - интерфейс источника данных URL
//	может реализовываться базой данных (Database) или структурами в оперативной памяти (Storage)
type Datasource interface {
	Insert(name, mType string, delta int64, value float64) error
	Get(name string) (mType string, delta int64, value float64, flg int)
	GetAll() []Metrics
	Close()
}

//	RowStorage - структура записи в хранилище метрик в оперативной памяти
//	используется для формирования структуры Storage и метода Storage.Insert
type MetricaRow struct {
	mType string
	value string
}

//	ErrEmptyNotAllowed - ошибка возникающая при попытке вставить пустое значение в любое поле структуры хранения URL
var ErrEmptyNotAllowed = errors.New("DataBase: empty value is not allowed")
