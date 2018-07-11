package main

import (
	"net/http"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

//type DBRow struct{
//	Name string
//	Value *sql.RawBytes
//}
//type DBColumn struct {
//	sql.ColumnType
//}

func RBtoString(val interface{}) string {
	return string(*(val.(*sql.RawBytes)))
}
func RBtoNullString(val interface{}) sql.NullString {
	s := string(*(val.(*sql.RawBytes)))
	if s != "" {
		return sql.NullString{
			String: s,
			Valid:  true,
		}
	} else {
		return sql.NullString{
			String: s,
			Valid:  false,
		}
	}
}
func RBtoInt(val interface{}) (int, error) {
	return strconv.Atoi(string(*(val.(*sql.RawBytes))))
}

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
	tables, err := mux.getTables()
	if err != nil {
		log.Print(err)
	}

	// GET / - список таблиц
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

	// GET  /table_name
	// POST /table_name/{id}
	// PUT  /table_name/{id}
	for _, table := range tables {
		route := "/"+ table+ "/"
		mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
			path := strings.Split(req.URL.Path, "/")
			table := path[0]
			if req.Method == http.MethodGet{
				data := mux.query(table, "1", "0")
				res := NewResponse()
				res.Response["records"] = data
				b, err := res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(b)

			} else if req.Method == http.MethodPost{

			} else if req.Method == http.MethodPut{

			}
		})
	}


	return mux, nil
}


func (h *DBHandler) query(table, limit, offset string)[]map[string]interface{} {
	q := "SELECT * FROM " + table + " LIMIT " + limit + " OFFSET " + offset + ";"
	log.Println(q)
	rows, err := h.db.Query(q)
	if err != nil {
		panic(err)
	}

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	colType, err := rows.ColumnTypes()
	if err != nil {
		panic(err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}
	var data []map[string]interface{}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			panic(err)
		}

		val := make(map[string]interface{}, len(cols))
		for i, v := range cols {
			switch colType[i].DatabaseTypeName() {
			case "INT":
				{
					val[v], _ = RBtoInt(vals[i])
				}
			case "VARCHAR", "TEXT":
				{
					if nullable, _ := colType[i].Nullable(); nullable {
						s := RBtoNullString(vals[i])
						if s.Valid {
							val[v] = s.String
						} else {
							val[v] = nil
						}
					} else {
						val[v] = RBtoString(vals[i])
					}
				}
			}
		}

		data = append(data, val)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		panic(err)
	}

	return data
}

func (h *DBHandler) getTables() ([]string, error) {
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