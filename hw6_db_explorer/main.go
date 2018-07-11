// тут лежит тестовый код
// менять вам может потребоваться только коннект к базе
package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"net/http"
)

var (
	// DSN это соединение с базой
	// вы можете изменить этот на тот который вам нужен
	// docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang2017 -d mysql
	// docker run -p 3306:3306 -v C:\Users\gg\go\src\github.com\germangorelkin\go_coursera2\hw6_db_explorer\:/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
	//DSN = "root@tcp(localhost:3306)/golang2017?charset=utf8"
	DSN = "root:1234@tcp(localhost:3306)/golang?charset=utf8"
)

func main() {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	// Exec
	//result, err := db.Exec("INSERT INTO items(`title`,`description`) VALUES (?, ?)",
	//	"test1", "desc test")
	//if err != nil {
	//	panic(err)
	//}
	//affected, err := result.RowsAffected()
	//if err != nil {
	//	panic(err)
	//}
	//lastID, err := result.LastInsertId()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Insert - RowsAffected", affected, "LastInsertId: ", lastID)
	//fmt.Printf("type: %T data: %+v\n", result, result)

	// QueryRow
	//row := db.QueryRow("SELECT id, title, updated, description FROM items WHERE id = ?", 1)
	//var id int
	//var title string
	//var updated sql.NullString
	//var description string
	//err = row.Scan(&id, &title, &updated, &description)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(id, title, updated, description)

	// Query
	//rows, err := db.Query("SELECT id, title, updated, description from items")
	//if err != nil {
	//	panic(err)
	//}
	//
	//cols, err := rows.Columns()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("columns: %s\n", cols)
	//
	//colType, err := rows.ColumnTypes()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("column type: %s\n\n", colType)
	//
	//for rows.Next() {
	//	var id int
	//	var title string
	//	var updated sql.NullString
	//	var description string
	//	err = rows.Scan(&id, &title, &updated, &description)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(id, title, updated, description)
	//}
	//rows.Close()
	//if err := rows.Err(); err != nil {
	//	panic(err)
	//}

	handler, err := NewDbExplorer(db)
	if err != nil {
		panic(err)
	}

	fmt.Println("starting server at :8082")
	http.ListenAndServe(":8082", handler)


	//se := make([]interface{}, 2)
	//se[0] = struct {}{
	//		"id":          1,
	//		"title":       "database/sql",
	//		"description": "Рассказать про базы данных",
	//		"updated":     "rvasily",
	//		}


	// Response
	//se := []map[string]interface{}{
	//	{
	//		"id":          1,
	//		"title":       "database/sql",
	//		"description": "Рассказать про базы данных",
	//		"updated":     "rvasily",
	//	},
	//	{
	//		"id":          2,
	//		"title":       "database/sql",
	//		"description": "Рассказать про базы данных",
	//		"updated":     "rvasily",
	//	},
	//}
	//
	//res := &DBResponse{}
	//res.Response = make(map[string]interface{})
	//res.Response["records"] = se
	//b, _ := res.Json()
	//fmt.Println(string(b))
}
