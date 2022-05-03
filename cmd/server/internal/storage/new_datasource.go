package storage

import (
	//	github.com/jackc/pgx/stdlib - драйвер PostgreSQL для доступа к БД с использованием пакета database/sql
	//	если хотим работать с БД напрямую, без database/sql надо использовать пакет - github.com/jackc/pgx/v4
	"errors"
	_ "github.com/jackc/pgx/stdlib"
)

// NewDatasource - функция конструктор, инициализирующая хранилище URL и интерфейсы работы с файлом, хранящим URL
func NewDatasource(databaseDSN, storeFile string, restoreOnStart bool) (strg Datasource, err error) {
	if databaseDSN != "" {
		return nil, errors.New("couldn't connect to DataBase")
	}
	s := Storage{Data: make([]Metrics, 0)}
	strg = &s

	//	опционально подключаем файл-хранилище метрик
	if storeFile != "" { //	если задан STORE_FILE, порождаем reader и writer для файла-хранилища
		fileReader, err = NewReader(storeFile)
		if err != nil { //	при ошибке создания reader, прерываем работу конструктора
			return nil, err
		}
		fileWriter, err = NewWriter(storeFile)
		if err != nil { //	при ошибке создания writer, прерываем работу конструктора
			return nil, err
		}

		//	если включена опция RESTORE - производим первичное заполнение хранилища метрик в оперативной памяти из файла
		if restoreOnStart == true {
			err := InitialFulfilment(&s)
			if err != nil { //	при ошибке первичного заполнения хранилища URL, прерываем работу конструктора
				return nil, err
			}
		}
	}

	return strg, nil //	если всё прошло ОК, то возращаем выбранный источник данных
}
