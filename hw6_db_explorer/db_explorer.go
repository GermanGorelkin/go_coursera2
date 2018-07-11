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
type DBColumn struct {
	Field string
	Type string
	Collation sql.NullString
	Null string
	Key string
	Default sql.NullString
	Extra string
	Privileges string
	Comment string
}

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

type sqlQuery string
func (q sqlQuery) Select(col ...string)sqlQuery {
	if len(col) == 0 {
		return sqlQuery("SELECT * ")
	}
	return sqlQuery("SELECT " + strings.Join(col, ",") + " ")
}
func (q sqlQuery) From(table string) sqlQuery {
	return sqlQuery(string(q) + " FROM " + table + " ")
}
func (q sqlQuery) WhereByPRIKey(columns []*DBColumn, val string) sqlQuery {
	if val == ""{
		return q
	}
	var key string
	for _, c := range columns {
		if c.Key == "PRI" {
			key = c.Field
			continue
		}
	}
	if key == "" {
		return q
	}
	return sqlQuery(string(q) + " WHERE " + key + "=" + val + " ")
}
func (q sqlQuery) Limit(limit string) sqlQuery {
	if _, err := strconv.Atoi(limit); err != nil {
		limit = "5"
	}
	return sqlQuery(string(q) + " LIMIT " + limit + " ")
}
func (q sqlQuery) Offset(offset string) sqlQuery {
	if _, err := strconv.Atoi(offset); err != nil {
		offset = "0"
	}
	return sqlQuery(string(q) + " OFFSET " + offset + " ")
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
		return mux, err
	}

	// GET / - список таблиц
	// или 404
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		res := NewResponse()

		if req.URL.Path != "/" || req.Method != http.MethodGet {
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
		columns, err := mux.getColumns(table)
		if err != nil {
			return mux, err
		}
		route := "/" + table + "/"
		mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
			path := strings.Split(req.URL.Path, "/")
			table := path[1]
			var id string
			if len(path)==3 {
				id = path[2]
			}

			if req.Method == http.MethodGet {
				limit := req.FormValue("limit")
				offset := req.FormValue("offset")
				var q sqlQuery
				q = q.Select().
					From(table).
					WhereByPRIKey(columns, id).
					Limit(limit).Offset(offset)
				log.Println(q)
				data, err := mux.query(q)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				res := NewResponse()
				if len(data)>0 {
					if id == "" {
						res.Response["records"] = data
					} else {
						res.Response["record"] = data[0]
					}
				} else{
					res.Error = "record not found"
					w.WriteHeader(http.StatusNotFound)
				}
				b, err := res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(b)

			} else if req.Method == http.MethodPost {

			} else if req.Method == http.MethodPut {

			}
		})
	}

	return mux, nil
}


func (h *DBHandler) query(q sqlQuery)([]map[string]interface{}, error) {
	var data []map[string]interface{}
	//q := "SELECT * FROM " + table +
	//	" LIMIT " + limit +
	//	" OFFSET " + offset
	//log.Println(q)
	rows, err := h.db.Query(string(q))
	if err != nil {
		return data, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return data, err
	}
	colType, err := rows.ColumnTypes()
	if err != nil {
		return data, err
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return data, err
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
		return data, err
	}

	return data, err
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

func (h *DBHandler) getColumns(table string) ([]*DBColumn, error) {
	rows, err := h.db.Query("SHOW FULL COLUMNS FROM `"+table+"`;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*DBColumn

	for rows.Next() {
		col := &DBColumn{}
		err = rows.Scan(
			&col.Field, &col.Type, &col.Collation,
			&col.Null, &col.Key, &col.Default, &col.Extra,
			&col.Privileges, &col.Comment)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}