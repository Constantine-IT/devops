package storage

import "errors"

//type gauge float64
//type counter int64

//	Datasource - интерфейс источника данных URL
//	может реализовываться базой данных (Database) или структурами в оперативной памяти (Storage)
type Datasource interface {
	Insert(name, mType, value string) error
	Get(name string) (mType, value string, flg int)
	GetAll() ([]MetricaValue, bool)
	Close()
}

//	RowStorage - структура записи в хранилище метрик в оперативной памяти
//	используется для формирования структуры Storage и метода Storage.Insert
type MetricaRow struct {
	mType string
	value string
}

//	RowStorage - структура записи в хранилище метрик в оперативной памяти
//	используется для формирования структуры Storage и метода Storage.Insert
type MetricaValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//	ErrEmptyNotAllowed - ошибка возникающая при попытке вставить пустое значение в любое поле структуры хранения URL
var ErrEmptyNotAllowed = errors.New("DataBase: empty value is not allowed")
