package main

import (
	"net/http"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"io/ioutil"
	"fmt"
	"reflect"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

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
func (c *DBColumn) GetType() string{
	switch {
	case strings.Contains(c.Type, "varchar"), c.Type == "text":
		{
			return "string"
		}
	case c.Type == "int":
		{
			return "int"
		}
	case c.Type == "float":
		{
			return "float"
		}
	}
	return ""
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
func (q sqlQuery) Select(col []string)sqlQuery {
	if len(col) == 0 {
		return sqlQuery("SELECT * ")
	}
	return sqlQuery(fmt.Sprintf("SELECT %s ", strings.Join(col, ",")))
}
func (q sqlQuery) From(table string) sqlQuery {
	return sqlQuery(fmt.Sprintf("%s FROM %s ", string(q), table))
}
func (q sqlQuery) UpdateTable(table string) sqlQuery {
	return sqlQuery(fmt.Sprintf("UPDATE %s ", table))
}
func (q sqlQuery) WhereByPRIKey(columns []*DBColumn, val string) sqlQuery {
	if val == ""{
		return q
	}
	key := getPRI(columns)
	if key == "" {
		return q
	}
	return sqlQuery(fmt.Sprintf("%s WHERE %s=%s ", string(q), key, val))
}
func (q sqlQuery) Limit(limit string) sqlQuery {
	if _, err := strconv.Atoi(limit); err != nil {
		limit = "5"
	}
	return sqlQuery(fmt.Sprintf("%s LIMIT %s ", string(q), limit))
}
func (q sqlQuery) Offset(offset string) sqlQuery {
	if _, err := strconv.Atoi(offset); err != nil {
		offset = "0"
	}
	return sqlQuery(fmt.Sprintf("%s OFFSET %s ;", string(q), offset))
}
func (q sqlQuery) Update(columns []*DBColumn, vals map[string]interface{}) (sqlQuery, error) {
	s := "SET "
	for _, col := range columns {
		name := col.Field
		v, ok := vals[name]

		if v != nil {
			if !strings.Contains(reflect.TypeOf(v).Name(), col.GetType()){
				return q, fmt.Errorf("field %s have invalid type", name)
			}
			//fmt.Println(reflect.TypeOf(v).Name(), col.GetType())
		}
		if col.Null == "NO" && ok && v==nil {
			return q, fmt.Errorf("field %s have invalid type", name)
		}

		//log.Println(name, v, col.Null, col.Key, ok)
		if col.Key == "PRI" && ok {
			return q, fmt.Errorf("field %s have invalid type", name)
		}
		if col.Key == "PRI" || !ok{
			continue
		}
		if col.Null == "YES" && ok && v==nil {
			v = "NULL"
		}

		if (strings.Contains(col.Type, "varchar") || col.Type == "text") && v != "NULL"{
			v = fmt.Sprintf("'%s'", v)
		}

		s = fmt.Sprintf("%s %s=%v,", s, name, v)
	}
	s = strings.TrimSuffix(s, ",")

	return sqlQuery(s), nil
}
func (q sqlQuery) Insert(table string, columns []*DBColumn, vals map[string]interface{}) (sqlQuery, error) {
	var scol string
	var sval string
	for _, col := range columns {
		name := col.Field
		v, ok := vals[name]

		//log.Println(name, v, col.Null, col.Key, ok)
		if col.Key == "PRI"{
			continue
		}
		if col.Null == "YES" && ok {
			v = "null"
		} else if col.Null == "YES" && !ok{
			continue
		}

		if strings.Contains(col.Type, "varchar") || col.Type == "text" {
			if v == nil {
				v = "''"
			} else {
				v = strings.Replace(v.(string), `'`,`\'`, 1)
				v = fmt.Sprintf("'%s'", v)
			}
		}

		scol = fmt.Sprintf("%s %s,", scol, name)
		sval = fmt.Sprintf("%s %v,", sval, v)
	}

	scol = strings.TrimSuffix(scol, ",")
	sval = strings.TrimSuffix(sval, ",")
	result := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, scol, sval)
	//log.Println(result)
	return sqlQuery(result), nil
}
func (q sqlQuery) Delete(table, id string, columns []*DBColumn) sqlQuery {
	key := getPRI(columns)
	if key == "" {
		return q
	}
	return sqlQuery(fmt.Sprintf("DELETE FROM %s WHERE %s=%s", table, key, id))
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


func getPRI(columns []*DBColumn)(field string) {
	for _, c := range columns {
		if c.Key == "PRI" {
			return c.Field
		}
	}
	return
}

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
			var cols []string
			for _,col := range columns{
				cols = append(cols, col.Field)
			}

			path := strings.Split(req.URL.Path, "/")
			table := path[1]
			var id string
			if len(path) == 3 {
				id = path[2]
			}

			if req.Method == http.MethodGet {
				//limit := req.FormValue("limit")
				//offset := req.FormValue("offset")
				var limit string
				if l, ok := req.URL.Query()["limit"]; ok{
					limit = l[0]
				}
				var offset string
				if l, ok := req.URL.Query()["offset"]; ok{
					offset = l[0]
				}
				//log.Println(limit, " ", offset)

				var q sqlQuery
				q = q.Select(cols).
					From(table).
					WhereByPRIKey(columns, id).
					Limit(limit).Offset(offset)
				log.Println(q)
				data, err := mux.query(q)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				res := NewResponse()
				if len(data) > 0 {
					if id == "" {
						res.Response["records"] = data
					} else {
						res.Response["record"] = data[0]
					}
				} else {
					res.Error = "record not found"
					w.WriteHeader(http.StatusNotFound)
				}
				b, err := res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				//log.Println(string(b))
				w.Write(b)

			} else if req.Method == http.MethodPost {
				res := NewResponse()
				b, err := ioutil.ReadAll(req.Body)
				if err != nil {
					//log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer req.Body.Close()

				var data map[string]interface{}
				json.Unmarshal(b, &data)

				var q sqlQuery
				q = q.UpdateTable(table)
				upd, err := q.Update(columns, data)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					res.Error = err.Error()
					b, _ := res.Json()
					w.Write(b)
					return
				}
				q = q + upd
				q = q.WhereByPRIKey(columns, id)
				//log.Println(q)

				_, eff, err := mux.exec(q)
				//log.Println(eff)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res.Response["updated"] = eff
				b, err = res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(b)

			} else if req.Method == http.MethodPut {
				b, err := ioutil.ReadAll(req.Body)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer req.Body.Close()

				var data map[string]interface{}
				json.Unmarshal(b, &data)

				var q sqlQuery
				q, err = q.Insert(table, columns, data)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				//log.Println(q)
				id, _, err := mux.exec(q)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				key := getPRI(columns)
				if key == "" {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				res := NewResponse()
				res.Response[key] = id
				b, err = res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(b)
			}else if req.Method == http.MethodDelete {
				var q sqlQuery
				q = q.Delete(table, id, columns)

				_, eff, err := mux.exec(q)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res := NewResponse()
				res.Response["deleted"] = eff
				b, err := res.Json()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(b)
			}
		})
	}

	return mux, nil
}

func (h *DBHandler) exec(q sqlQuery) (id int64, eff int64, err error) {
	res, err := h.db.Exec(string(q))
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	eff, err = res.RowsAffected()
	if err != nil {
		return
	}
	return
}

func (h *DBHandler) query(q sqlQuery)([]map[string]interface{}, error) {
	var data []map[string]interface{}

	rows, err := h.db.Query(string(q))
	if rows != nil {
		defer rows.Close()
	}
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
			b := *(vals[i].(*sql.RawBytes))

			if strings.Contains(colType[i].DatabaseTypeName(), "INT") {
				if len(b) > 0 {
					val[v], _ = strconv.Atoi(string(b))
				} else if nullable, _ := colType[i].Nullable(); nullable {
					val[v] = nil
				} else {
					val[v] = 0
				}

			} else if strings.Contains(colType[i].DatabaseTypeName(), "FLOAT") {
				if len(b) > 0 {
					val[v], _ = strconv.ParseFloat(string(b), 64)
				} else if nullable, _ := colType[i].Nullable(); nullable {
					val[v] = nil
				} else {
					val[v] = 0
				}
			} else {
				if len(b) > 0 {
					val[v] = string(b)
				} else if nullable, _ := colType[i].Nullable(); nullable {
					val[v] = nil
				} else {
					val[v] = ""
				}
			}
		}

		data = append(data, val)
	}

	if err := rows.Err(); err != nil {
		return data, err
	}

	return data, err
}

func (h *DBHandler) getTables() ([]string, error) {
	rows, err := h.db.Query("SHOW TABLES;")
	if rows != nil{
		defer rows.Close()
	}
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
	if rows != nil{
		defer rows.Close()
	}
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
