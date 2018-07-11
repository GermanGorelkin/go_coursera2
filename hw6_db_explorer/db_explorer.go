package main

import (
	"net/http"
	"database/sql"
	"encoding/json"
	"log"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DBResponse struct {
	Response map[string]interface{} `json:"response,omitempty"`
	Error string `json:"error,omitempty"`
}
func (r *DBResponse) Json() ([]byte, error) {
	return json.Marshal(r)
}
func NewResponse() *DBResponse {
	return &DBResponse{Response: make(map[string]interface{})}
}

type DBHandler struct{
	http.ServeMux
	db *sql.DB
}
//func (mux *DBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	h, pattern := mux.Handler(r)
//	log.Println(h, pattern)
//	h.ServeHTTP(w, r)
//}

func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{db: db}
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	mux := NewDBHandler(db)

	// all table
	tables, err := mux.GetTables()
	if err != nil {
		log.Print(err)
	}

	// список таблиц для GET /
	// или 404
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		res := NewResponse()

		if req.URL.Path != "/" || req.Method != http.MethodGet{
			w.WriteHeader(http.StatusNotFound)
			res.Error = "unknown table"
			b, err := res.Json()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			return
		}

		res.Response["tables"] = tables
		b, err := res.Json()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	return mux, nil
}

func (h *DBHandler) GetTables() ([]string, error){
	rows, err := h.db.Query("SHOW TABLES;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}