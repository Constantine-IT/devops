package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
	"log"
	"os"
	"time"
)

// NewDatasource - функция конструктор, инициализирующая хранилище URL и интерфейсы работы с файлом, хранящим URL
func NewDatasource(databaseDSN, storeFile string, storeInterval time.Duration, restoreOnStart bool) (strg Datasource, err error) {
	//	Приоритетность в использовании ресурсов хранения информации URL (по убыванию приоритета):
	//	1.	Внешняя база данных, параметры соединения с которой задаются через DATABASE_DSN
	//	2.	Если БД не задана, то используем файловое хранилище (задаваемое через STORE_FILE) и оперативную память
	//	3.	Если не заданы ни БД, ни файловое хранилище, то работаем только с оперативной памятью - структура storage.Storage

	if databaseDSN != "" { //	если задана переменная среды DATABASE_DSN
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
	} else { //	если база данных не задана, то работаем со структурой хранения в оперативной памяти
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
			if restoreOnStart {
				err := InitialFulfilment(&s)
				log.Println("Initial load metrics from file - SUCCESS")
				if err != nil { //	при ошибке первичного заполнения хранилища URL, прерываем работу конструктора
					return nil, err
				}
			}
			go func() { //	запускаем отдельный воркер - записи метрик в файл на периодической основе
				if storeInterval <= 0 { //	минимальный интервал сброса дампа метрик в файл - 1 секунда
					storeInterval = 1
				}
				// создаём тикер, подающий раз в StoreInterval секунд, сигнал на запись метрик в файл
				fileWriteTicker := time.NewTicker(storeInterval * time.Second)
				defer fileWriteTicker.Stop()

				// запускаем слежение за каналами тикера записи в файл
				for {
					<-fileWriteTicker.C
					//	пишем метрики в файл
					if err := DumpToFile(&s); err != nil {
						log.Println("SERVER metrics collector unable to write to file - (code 1) SHUTDOWN")
						os.Exit(1)
					}
					log.Println("All metrics were written to file:", storeFile)
				}

			}()
		}
	}

	return strg, nil //	если всё прошло ОК, то возращаем выбранный источник данных
}
