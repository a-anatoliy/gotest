package mysql

import (
	"frmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func main () {
	
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/mytestdb")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	fmt.Println("Working with MySQL")
}