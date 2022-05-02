package storage

import (
	//	github.com/jackc/pgx/stdlib - драйвер PostgreSQL для доступа к БД с использованием пакета database/sql
	//	если хотим работать с БД напрямую, без database/sql надо использовать пакет - github.com/jackc/pgx/v4
	"errors"
	_ "github.com/jackc/pgx/stdlib"
)

// NewDatasource - функция конструктор, инициализирующая хранилище URL и интерфейсы работы с файлом, хранящим URL
func NewDatasource(dbDSN, file string) (strg Datasource, err error) {
	if dbDSN != "" || file != "" {
		return nil, errors.New("couldn't connect to DataBase or File")
	}
	s := Storage{Data: make([]Metrics, 0)}
	strg = &s
	return strg, nil //	если всё прошло ОК, то возращаем выбранный источник данных для URL
}
