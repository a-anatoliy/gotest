package mysql

import (
  "fmt"
  "database/sql"

  _ "github.com/go-sql-driver/mysql"
)

type User struct {
  Name string `json:"name"`
  Age uint16 `json:"age"`
}

func main () {
  
  db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/mytestdb")
  if err != nil {
    panic(err)
  }

  defer db.Close()

  fmt.Println("Connected to MySQL")
  // fmt.Println("Inserting initial data")

  // insert, err := db.Query("insert into `mytestdb` (`name`,`age`) values ('Marty',28)")
  // insert, err := db.Query("insert into `mytestdb` (`name`,`age`) values ('Wandy',27)")
  // if err != nil {
  //   panic(err)
  // }

  // defer insert.Close()

  result, err := db.Query("select name,age from `mytestdb`")
  if err != nil {
    panic(err)
  }

  for result.Next() {
    var user User
    err = result.Scan(&user.Name, &user.Age)
    if err != nil {
      panic(err)
    }

    // fmt.Sprintf("User: %s with age: %d", user.Name, user.Age)
    // fmt.Printf("User: %s with age: %d", user.Name, user.Age)
    fmt.Println(fmt.Sprintf("User: %s with age: %d", user.Name, user.Age))
  }


  defer result.Close()


}