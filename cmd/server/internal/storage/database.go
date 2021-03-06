package storage

import (
	"errors"
	"io"
	"log"

	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
)

//	Методы для работы с данными в структуре базы данных - Database

// Insert - Метод для сохранения метрик
func (d *Database) Insert(name, mType string, delta int64, value float64) error {
	//	пустые значения к вставке в хранилище не допускаются
	if name == "" || mType == "" {
		return ErrEmptyNotAllowed
	}

	//	проверяем, есть ли метрика с таким именем в нашей базе
	Type, Delta, _, Flag := d.Get(name)
	if Flag == 1 { //	если метрика с таким именем уже содержится в нашей базе данных
		if Type == "counter" { //	для метрик типа counter новое значение прибавляется к старому
			delta = delta + Delta
		}
		//	начинаем тразакцию
		tx, err := d.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback() //	при ошибке выполнения - откатываем транзакцию

		//	готовим SQL-statement для обновления значений метрики в базе данных
		stmt, err := tx.Prepare(`update "metrics" set "type" = $1, "delta" = $2, "value" = $3 where "name" = $4`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		//	 запускаем SQL-statement на исполнение передавая в него параметры метрики
		if _, err := stmt.Exec(mType, delta, value, name); err != nil {
			return err
		}

		return tx.Commit() //	при успешном выполнении вставки - фиксируем транзакцию
	}
	if Flag == 0 { //	если метрики с таким именем в нашей базе данных НЕТ
		//	начинаем тразакцию
		tx, err := d.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback() //	при ошибке выполнения - откатываем транзакцию

		//	готовим SQL-statement для вставки в базу новой метрики
		stmt, err := tx.Prepare(`insert into "metrics" ("name", "type", "delta", "value") values ($1, $2, $3, $4)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		//	 запускаем SQL-statement на исполнение
		if _, err := stmt.Exec(name, mType, delta, value); err != nil {
			return err
		}

		return tx.Commit() //	при успешном выполнении вставки - фиксируем транзакцию
	}
	return nil
}

// Get - метод для нахождения значения метрики по её имени
func (d *Database) Get(name string) (mType string, delta int64, value float64, flg int) {
	//	готовим запрос в базу данных и запускаем его на исполнение
	stmt := `select "type", "delta", "value" from "metrics" where "name" = $1`
	err := d.DB.QueryRow(stmt, name).Scan(&mType, &delta, &value)
	if errors.Is(err, sql.ErrNoRows) { //	если запрос не вернул ни одной строки
		return "", 0, 0, 0 //	если метрика с искомым имененм НЕ найдена, возвращаем flag=0
	}
	return mType, delta, value, 1 //	если метрика найдена, возвращаем flag=1
}

//	GetAll - метод возсвращает всё содержимое хранилища метрик
func (d *Database) GetAll() (result []Metrics) {

	metrica := Metrics{}

	//	готовим запрос на выборку всех строк из таблицы метрик и запускаем его на исполнение
	stmt := `select "name", "type", "delta", "value" from "metrics"`
	rows, err := d.DB.Query(stmt)
	if err != nil || rows.Err() != nil {
		return nil
	}
	defer rows.Close()
	//	перебираем по одной все строки выборки, добавляя метрики в исходящий срез
	for rows.Next() {
		err := rows.Scan(&metrica.ID, &metrica.MType, &metrica.Delta, &metrica.Value)
		if err != nil {
			return nil
		}
		result = append(result, metrica)
	}
	return result
}

//	Close - метод закрытия структур хранения
func (d *Database) Close() {
	//	при остановке сервера закрываем connect с базой данных
	d.DB.Close()
}

//	InitialFulfilment - метод первичного заполнения хранилища метрик из файла-хранилища, при старте сервера
func (d *Database) InitialFulfilment() {
	for {
		metrica, err := fileReader.Read() //	считываем записи по одной из файла-хранилища

		if errors.Is(err, io.EOF) { //	когда дойдем до конца файла - выходим из цикла чтения
			log.Println("initial load metrics from file - SUCCESS")
			break
		}
		if err != nil { //	при ошибке чтения метрики - пропускаем её и читаем файл дальше
			log.Println("file read error due to InitialFulfilment process")
			continue
		}

		d.Insert(metrica.ID, metrica.MType, metrica.Delta, metrica.Value)
	}
}
