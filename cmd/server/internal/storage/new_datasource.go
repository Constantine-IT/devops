package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
	"log"
	"os"
	"time"
)

// NewDatasource - функция конструктор, инициализирующая источник данных для метрик и интерфейсы работы с файлом-хранилищем
func NewDatasource(databaseDSN, storeFile string, storeInterval time.Duration, restoreOnStart bool) (strg Datasource, err error) {
	//	Приоритетность в использовании ресурсов хранения информации URL (по убыванию приоритета):
	//	1.	Внешняя база данных, параметры соединения с которой задаются через DATABASE_DSN
	//	2.	Если БД не задана, то используем файловое хранилище (задаваемое через STORE_FILE) и оперативную память
	//	3.	Если не заданы ни БД, ни файловое хранилище, то работаем только с оперативной памятью - структура storage.Storage

	if databaseDSN != "" { //	если задана переменная DATABASE_DSN
		var err error
		var d Database
		//	открываем connect с базой данных PostgreSQL 10+
		d.DB, err = sql.Open("pgx", databaseDSN)
		if err != nil { //	при ошибке открытия, прерываем работу конструктора
			return nil, err
		}
		//	тестируем доступность базы данных
		if err := d.DB.Ping(); err != nil { //	если база недоступна, прерываем работу конструктора
			return nil, err
		} else { //	если база данных доступна - создаём в ней структуры хранения

			//	готовим SQL-statement для создания таблицы metrics, если её не существует
			stmt := `create table if not exists "metrics" (
									"name" TEXT constraint name_pk primary key not null,
			   						"type" TEXT not null,
			   						"delta" BIGINT not null,
			                        "value" DOUBLE PRECISION not null)`
			_, err := d.DB.Exec(stmt)
			if err != nil { //	при ошибке в создании структур хранения в базе данных, прерываем работу конструктора
				return nil, err
			}
		}
		//	если всё прошло успешно, возвращаем в качестве источника данных - базу данных
		strg = &Database{DB: d.DB}
		//	если задан STORE_FILE и включена опция RESTORE - загружаем из файла список метрик
		if restoreOnStart && storeFile != "" {
			fileReader, err = NewReader(storeFile) //	 порождаем reader для файла-хранилища
			if err != nil {                        //	при ошибке создания reader, прерываем работу конструктора
				return nil, err
			}
			strg.InitialFulfilment() //	производим первичное заполнение хранилища метрик из файла
			// закрываем reader для файла-хранилища
			fileReader.Close()
		}
	} else { //	если база данных не задана, то работаем со структурой хранения в оперативной памяти
		s := Storage{Data: make([]Metrics, 0)}
		strg = &s
		if storeFile != "" { //	если задан STORE_FILE - опционально подключаем файл-хранилище метрик

			if restoreOnStart { //	если включена опция RESTORE - загружаем из файла список метрик
				fileReader, err = NewReader(storeFile) //	 порождаем reader для файла-хранилища
				if err != nil {                        //	при ошибке создания reader, прерываем работу конструктора
					return nil, err
				}
				strg.InitialFulfilment() //	производим первичное заполнение хранилища метрик из файла
				// закрываем reader для файла-хранилища
				fileReader.Close()
			}

			fileWriter, err = NewWriter(storeFile) //	порождаем writer для файла-хранилища
			if err != nil {                        //	при ошибке создания writer, прерываем работу конструктора
				return nil, err
			}

			//	запускаем отдельный воркер - записи метрик в файл на периодической основе
			go func() {
				if storeInterval < 1*time.Second { //	минимальный интервал сброса дампа метрик в файл - 1 секунда
					storeInterval = 1 * time.Second
				}
				// создаём тикер, подающий раз в StoreInterval секунд, сигнал на запись метрик в файл
				fileWriteTicker := time.NewTicker(storeInterval)
				defer fileWriteTicker.Stop()

				for { // запускаем слежение за каналами тикера записи в файл
					<-fileWriteTicker.C
					//	пишем метрики в файл
					if err := DumpToFile(&s); err != nil {
						log.Println("SERVER metrics collector unable to write to file - (code 1) SHUTDOWN")
						os.Exit(1)
					}
				}

			}()
		}
	}
	return strg, nil //	если всё прошло ОК, то возращаем выбранный источник данных
}
