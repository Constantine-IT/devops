package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
)

//	Database - структура хранилища URL, обертывающая пул подключений к базе данных
type Database struct {
	DB *sql.DB
}

// Insert - Метод для сохранения метрик
func (d *Database) Insert(name, mType string, delta int64, value float64) error {
	//	пустые значения к вставке в хранилище не допускаются
	if name == "" || mType == "" {
		return ErrEmptyNotAllowed
	}

	return nil
}

// Get - метод для нахождения значения метрики
func (d *Database) Get(name string) (mType string, delta int64, value float64, flg int) {
	return "", 0, 0, 0 //	если метрика с искомым имененм НЕ найдена, возвращаем flag=0
}

func (d *Database) GetAll() []Metrics {
	return nil
}

//	DumpToFile - сбрасывает все метрики в файловое хранилище
func (d *Database) DumpToFile() error {
	return nil
}

//	Close - метод, закрывающий reader и writer для файла-хранилища URL, а также connect к базе данных
func (d *Database) Close() error {
	//	при остановке сервера connect к базе данных
	err := d.DB.Close()
	if err != nil {
		return err
	}

	return nil
}
