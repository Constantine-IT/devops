package handlers

import (
	"compress/gzip"
	"net/http"
)

//	DecompressGZIP - middleware, распаковывающая тело сжатых GZIP запросов
func (app *Application) DecompressGZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` { //	если входящий пакет сжат GZIP
			gz, err := gzip.NewReader(r.Body) //	изготавливаем reader-декомпрессор GZIP
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				app.ErrorLog.Println("Request Body decompression error: " + err.Error())
				return
			}
			r.Body = gz //	подменяем стандартный reader из Request на декомпрессор GZIP
			defer gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}
