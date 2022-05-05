package handlers

import (
	"net/http"

	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
)

//	PingDataBaseHandler - обработчик проверки доступности базы данных через GET /ping
func (app *Application) PingDataBaseHandler(w http.ResponseWriter, r *http.Request) {
	switch value := app.Datasource.(type) {
	//	проверяем тип источника данных - Datasource
	case *storage.Database: //	если интерфейс источника данных Datasource реализован базой данных - Database
		err := value.DB.Ping() //	проверяем доступность базы данных
		if err != nil {        //	если база данных недоступна, то отвечаем с http.StatusInternalServerError
			http.Error(w, err.Error(), http.StatusInternalServerError)
			app.ErrorLog.Println("Try to PING database error: " + err.Error())
		} else { //	если база данных пингуется, то отвечаем со статусом http.StatusOK
			http.Error(w, http.StatusText(200), http.StatusOK)
		}
	default: //	если интерфейс источника данных Datasource реализован иной структурой
		http.Error(w, "DataBase environment variable wasn't set", http.StatusInternalServerError)
		app.ErrorLog.Println("Attempt to PING database, that wasn't set in server configuration")
		return
	}

}
