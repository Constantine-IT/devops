package storage

import "errors"

//	Datasource - интерфейс источника данных URL
//	может реализовываться базой данных (Database) или структурами в оперативной памяти (Storage)
type Datasource interface {
	Insert(name, mType, value string) error
	Get(name string) (mType, value string, flg int)
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
