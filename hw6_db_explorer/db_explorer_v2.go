package main
/*
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type DbExplorer struct {
	http.ServeMux
	db *sql.DB
}
func (d *DbExplorer) loadDbInfo() tables {
	tbl := make(tables)

	tblNames, err := d.getTables()
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range tblNames {
		cols, err := d.getColumns(name)
		if err != nil {
			log.Fatal(err)
		}
		tbl[name] = cols
	}
	return tbl
}
func (d *DbExplorer) getColumns(table string) ([]*column, error) {
	rows, err := d.db.Query("SHOW FULL COLUMNS FROM `"+table+"`;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*column

	for rows.Next() {
		col := &column{}
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
func (d *DbExplorer) getTables() ([]string, error) {
	rows, err := d.db.Query("SHOW TABLES;")
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

// -----
func (d *DbExplorer) executeQuery(query string) ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	rows, err := d.db.Query(query)
	if err != nil {
		return data, err
	}
	defer rows.Close()

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
					val[v], _ = strconv.Atoi(string(*(vals[i].(*sql.RawBytes))))
				}
			case "VARCHAR", "TEXT":
				{
					if nullable, _ := colType[i].Nullable(); nullable {
						s := string(*(vals[i].(*sql.RawBytes)))
						if s != "" {
							val[v] = s
						} else {
							val[v] = nil
						}
					} else {
						val[v] = string(*(vals[i].(*sql.RawBytes)))
					}
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

func (d *DbExplorer) getAll(limit, offset int64, cols []*column, tblName string) ([]map[string]interface{}, error) {
	colNames := make([]string, 0)
	for _, col := range cols {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s LIMIT %d OFFSET %d`, columns, tblName, limit, offset)
	resp, err := d.executeQuery(query)

	return resp, err
}
func (d *DbExplorer) getByID(id int, cols []*column, tblName string) ([]map[string]interface{}, error) {
	colNames := make([]string, 0)
	for _, col := range cols {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = %d`, columns, tblName, cols[0].Field, id)
	resp, err := d.executeQuery(query)

	return resp, err
}

func (d *DbExplorer) insert(data map[string]interface{}, cols []*column, tableName string) (int64, error) {
	colNames := make([]string, 0)
	values := make([]interface{}, 0)

	for i := 1; i < len(cols); i++ {
		colNames = append(colNames, cols[i].Field)

		val, ok := data[cols[i].Field]
		if !ok {
			if cols[i].Null=="YES" {
				val = nil
			} else {
				val = ""
			}
		}

		values = append(values, val)
	}

	columnsStr := strings.Join(colNames, ", ")
	placeholders := "?" + strings.Repeat(", ?", len(values)-1)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, placeholders)
	res, err := d.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func (d *DbExplorer) update(id int, data map[string]interface{}, columns []*column, tableName string) (int64, error) {
	setStmts := make([]string, 0)
	values := make([]interface{}, 0)

	for k, v := range data {
		setStmts = append(setStmts, fmt.Sprintf("%v = ?", k))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %d", tableName, strings.Join(setStmts, ", "), columns[0].Field, id)
	res, err := d.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

func (d *DbExplorer) delete(id int, columns []*column, tableName string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", tableName, columns[0].Field)

	res, err := d.db.Exec(query, id)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}
//----

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	mux := &DbExplorer{db: db}
	tbl := mux.loadDbInfo()

	// GET / - список таблиц
	// или 404
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" || req.Method != http.MethodGet {
			writeResponseJSON(w, http.StatusNotFound, "", nil, "unknown table")
			return
		}
		tblNames, err := mux.getTables()
		if err != nil {
			log.Fatal(err)
		}
		writeResponseJSON(w, http.StatusOK, "tables", tblNames, "")
	})

	// GET  /table_name
	// POST /table_name/{id}
	// PUT  /table_name/{id}
	for table := range tbl {
		route := "/" + table + "/"
		mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
			path := strings.Split(req.URL.Path, "/")
			table := path[1]
			var id string
			if len(path) == 3 {
				id = path[2]
			}
			cols := tbl[table]

			if req.Method == http.MethodGet {
				if id != "" {
					id, _ := strconv.Atoi(id)
					resp, err := mux.getByID(id, cols, table)
					if err != nil {
						writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
						return
					}

					if len(resp) == 0 {
						writeResponseJSON(w, http.StatusNotFound, "", nil, "record not found")
						return
					}
					writeResponseJSON(w, http.StatusOK, "record", resp[0], "")
				} else {
					limit, err := strconv.ParseInt(req.URL.Query().Get("limit"), 0, 32)
					if err != nil {
						limit = 5
					}
					offset, err := strconv.ParseInt(req.URL.Query().Get("offset"), 0, 32)
					if err != nil {
						offset = 0
					}

					resp, err := mux.getAll(limit, offset, cols, table)
					if err != nil {
						log.Println(err)
						writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
						return
					}
					writeResponseJSON(w, http.StatusOK, "records", resp, "")
				}

			} else if req.Method == http.MethodPost {
				id, _ := strconv.Atoi(id)
				data := make(map[string]interface{})
				err := json.NewDecoder(req.Body).Decode(&data)
				if err != nil {
					writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
					return
				}

				err = validateFields(data, cols)
				if err != nil {
					writeResponseJSON(w, http.StatusBadRequest, "", nil, err.Error())
					return
				}

				rowsAffected, err := mux.update(id, data, cols, table)
				if err != nil {
					writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
					return
				}

				writeResponseJSON(w, http.StatusOK, "updated", rowsAffected, "")

			} else if req.Method == http.MethodPut {
				data := make(map[string]interface{})
				err := json.NewDecoder(req.Body).Decode(&data)
				if err != nil {
					writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
					return
				}

				lastID, err := mux.insert(data, cols, table)
				if err != nil {
					writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
					return
				}

				writeResponseJSON(w, http.StatusOK, cols[0].Field, lastID, "")

			} else if req.Method == http.MethodDelete {
				id, _ := strconv.Atoi(id)
				rowsAffected, err := mux.delete(id, cols, table)
				if err != nil {
					writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
					return
				}

				writeResponseJSON(w, http.StatusOK, "deleted", rowsAffected, "")
			}
		})
	}

	return mux, nil
}

// models
type tables map[string][]*column

type column struct {
	Field string
	Type string
	Collation *string
	Null string
	Key string
	Default *string
	Extra string
	Privileges string
	Comment string
}

func validateFields(data map[string]interface{}, columns []*column) error {
	if len(data) == 1 {
		_, ok := data["id"]
		if ok {
			return errors.New("field id have invalid type")
		}
	}

	for _, column := range columns {
		val, ok := data[column.Field]
		if !ok {
			continue
		}

		if val == nil && column.Null=="YES" {
			return nil
		}

		if val == nil && column.Null=="NO" {
			return fmt.Errorf("field %s have invalid type", column.Field)
		}

		if !compareTypes(column.Type, reflect.TypeOf(val).Name()) {
			return fmt.Errorf("field %s have invalid type", column.Field)
		}
	}

	return nil
}
func compareTypes(colType, fieldType string) bool {
	switch fieldType {
	case "string":
		if colType == "varchar(255)" || colType == "text" {
			return true
		}

	default:
		return false
	}

	return false
}

// response
type response struct {
	Error    string                 `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}
func writeResponseJSON(w http.ResponseWriter, status int, dataName string, data interface{}, errorText string) {
	w.Header().Set("Content-Type", "application/json")
	var resp response
	if errorText != "" {
		resp = response{
			Error: errorText,
		}
	} else {
		resMap := make(map[string]interface{})
		resMap[dataName] = data
		resp = response{
			Response: resMap,
		}
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
	} else {
		w.WriteHeader(status)
		w.Write(jsonResp)
	}
}
*/